package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/vibe-dashboard/vibe-dashboard/sources"
)

type Store struct {
	db *sql.DB
}

func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(home, ".vibe-dashboard")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dir, "vibe.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open store db: %w", err)
	}

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
		`CREATE TABLE IF NOT EXISTS budget_alerts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			threshold_cents INTEGER NOT NULL,
			action TEXT DEFAULT 'warn',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_agent ON sessions(agent)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project)`,
		`CREATE INDEX IF NOT EXISTS idx_file_changes_session ON file_changes(session_id)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) SaveSession(ses sources.Session) error {
	_, err := s.db.Exec(`
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

func (s *Store) SaveFileChanges(sessionID string, changes []sources.FileChange) error {
	tx, err := s.db.Begin()
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

func (s *Store) GetSessions(agent string, limit int) ([]sources.Session, error) {
	if limit <= 0 {
		limit = 50
	}

	var rows *sql.Rows
	var err error
	if agent != "" {
		rows, err = s.db.Query(`
			SELECT id, agent, project, status, start_time, duration_seconds,
				   cost, input_tokens, output_tokens, cache_tokens, cache_hit_rate
			FROM sessions WHERE agent = ? ORDER BY start_time DESC LIMIT ?
		`, agent, limit)
	} else {
		rows, err = s.db.Query(`
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
		err := rows.Scan(&ses.ID, &ses.Agent, &ses.Project, &ses.Status, &startStr,
			&ses.Duration, &ses.Cost, &ses.InputTokens, &ses.OutputTokens,
			&ses.CacheTokens, &ses.CacheHitRate)
		if err != nil {
			continue
		}
		ses.StartTime, _ = time.Parse("2006-01-02T15:04:05Z", startStr)
		sessions = append(sessions, ses)
	}
	return sessions, nil
}

func (s *Store) GetAggregatedCost(hours int, agent string) (float64, error) {
	var row *sql.Row
	if agent != "" {
		row = s.db.QueryRow(`
			SELECT COALESCE(SUM(cost), 0) FROM sessions 
			WHERE start_time > datetime('now', ?) AND agent = ?
		`, fmt.Sprintf("-%d hours", hours), agent)
	} else {
		row = s.db.QueryRow(`
			SELECT COALESCE(SUM(cost), 0) FROM sessions 
			WHERE start_time > datetime('now', ?)
		`, fmt.Sprintf("-%d hours", hours))
	}

	var total float64
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
