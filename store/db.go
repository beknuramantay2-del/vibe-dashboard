package store

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"github.com/vibe-dashboard/vibe-dashboard/sources"
)

// Store persists aggregated session data to a local SQLite database.
type Store struct {
	mu sync.RWMutex
	db *sql.DB
}

// NewStore creates a store at ~/.vibe-dashboard/vibe.db.
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(home, ".vibe-dashboard")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dir, "vibe.db")
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open store db: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping store db: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate store: %w", err)
	}
	return s, nil
}

func (s *Store) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			agent TEXT NOT NULL,
			project TEXT,
			status TEXT DEFAULT 'active',
			start_time DATETIME,
			duration_seconds REAL DEFAULT 0,
			cost REAL DEFAULT 0,
			input_tokens INTEGER DEFAULT 0,
			output_tokens INTEGER DEFAULT 0,
			cache_tokens INTEGER DEFAULT 0,
			cache_hit_rate REAL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS file_changes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL REFERENCES sessions(id),
			file_path TEXT NOT NULL,
			additions INTEGER DEFAULT 0,
			deletions INTEGER DEFAULT 0,
			diff_content TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_agent ON sessions(agent)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_start_time ON sessions(start_time)`,
		`CREATE INDEX IF NOT EXISTS idx_file_changes_session ON file_changes(session_id)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) getDB() (*sql.DB, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, ErrClosed
	}
	return s.db, nil
}

// SaveSession upserts a session record.
func (s *Store) SaveSession(ses sources.Session) error {
	db, err := s.getDB()
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		INSERT INTO sessions (id, agent, project, status, start_time, duration_seconds, cost, input_tokens, output_tokens, cache_tokens, cache_hit_rate)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			status = excluded.status,
			duration_seconds = excluded.duration_seconds,
			cost = excluded.cost,
			input_tokens = excluded.input_tokens,
			output_tokens = excluded.output_tokens,
			cache_tokens = excluded.cache_tokens,
			cache_hit_rate = excluded.cache_hit_rate
	`, ses.ID, ses.Agent, ses.Project, ses.Status, ses.StartTime.UTC(),
		ses.Duration.Seconds(), ses.Cost, ses.InputTokens, ses.OutputTokens,
		ses.CacheTokens, ses.CacheHitRate)
	return err
}

// SaveFileChanges saves a batch of file changes for a session.
func (s *Store) SaveFileChanges(sessionID string, changes []sources.FileChange) error {
	db, err := s.getDB()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO file_changes (session_id, file_path, additions, deletions)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range changes {
		if _, err := stmt.Exec(sessionID, c.Path, c.Additions, c.Deletions); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetSessions retrieves sessions filtered by agent, with a limit.
func (s *Store) GetSessions(agent string, limit int) ([]sources.Session, error) {
	db, err := s.getDB()
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}

	var rows *sql.Rows
	if agent != "" {
		rows, err = db.Query(`
			SELECT id, agent, project, status, start_time, duration_seconds,
			       cost, input_tokens, output_tokens, cache_tokens, cache_hit_rate
			FROM sessions WHERE agent = ? ORDER BY start_time DESC LIMIT ?
		`, agent, limit)
	} else {
		rows, err = db.Query(`
			SELECT id, agent, project, status, start_time, duration_seconds,
			       cost, input_tokens, output_tokens, cache_tokens, cache_hit_rate
			FROM sessions ORDER BY start_time DESC LIMIT ?
		`, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []sources.Session
	for rows.Next() {
		var ses sources.Session
		var startStr string
		var durationSec float64
		err := rows.Scan(&ses.ID, &ses.Agent, &ses.Project, &ses.Status, &startStr,
			&durationSec, &ses.Cost, &ses.InputTokens, &ses.OutputTokens,
			&ses.CacheTokens, &ses.CacheHitRate)
		if err != nil {
			log.Printf("store: row scan: %v", err)
			continue
		}

		// Parse start time with multiple format attempts
		for _, layout := range []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05",
		} {
			if parsed, e := time.Parse(layout, startStr); e == nil {
				ses.StartTime = parsed
				break
			}
		}

		ses.Duration = time.Duration(durationSec * float64(time.Second))
		sessions = append(sessions, ses)
	}
	return sessions, rows.Err()
}

// GetAggregatedCost returns total cost, optionally filtered by hours and agent.
func (s *Store) GetAggregatedCost(hours int, agent string) (float64, error) {
	db, err := s.getDB()
	if err != nil {
		return 0, err
	}

	var query string
	var args []interface{}

	switch {
	case agent != "" && hours > 0:
		query = `SELECT COALESCE(SUM(cost), 0) FROM sessions WHERE start_time > datetime('now', '-' || ? || ' hours') AND agent = ?`
		args = []interface{}{hours, agent}
	case hours > 0:
		query = `SELECT COALESCE(SUM(cost), 0) FROM sessions WHERE start_time > datetime('now', '-' || ? || ' hours')`
		args = []interface{}{hours}
	case agent != "":
		query = `SELECT COALESCE(SUM(cost), 0) FROM sessions WHERE agent = ?`
		args = []interface{}{agent}
	default:
		query = `SELECT COALESCE(SUM(cost), 0) FROM sessions`
	}

	var total float64
	if err := db.QueryRow(query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// ErrClosed is returned when operating on a closed store.
var ErrClosed = errors.New("store already closed")

// Close closes the underlying database connection.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return ErrClosed
	}
	db := s.db
	s.db = nil
	return db.Close()
}
