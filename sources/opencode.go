package sources

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// OpenCodeReader reads session data from the OpenCode SQLite database.
type OpenCodeReader struct {
	db       *sql.DB
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewOpenCodeReader() (*OpenCodeReader, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}

	paths := []string{
		filepath.Join(home, ".opencode", "opencode.db"),
		filepath.Join(home, ".local", "share", "opencode", "opencode.db"),
		filepath.Join(home, "AppData", "Roaming", "opencode", "opencode.db"),
	}

	var db *sql.DB
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			db, err = sql.Open("sqlite", p+"?mode=ro")
			if err != nil {
				continue
			}
			db.SetMaxOpenConns(1)
			if err = db.Ping(); err == nil {
				break
			}
			db.Close()
			db = nil
		}
	}

	if db == nil {
		return nil, fmt.Errorf("opencode database not found in any known path")
	}

	r := &OpenCodeReader{
		db:       db,
		sessions: make(map[string]*Session),
	}
	return r, nil
}

func (o *OpenCodeReader) Name() string { return "OpenCode" }

// Refresh reloads sessions from the SQLite database.
func (o *OpenCodeReader) Refresh() error {
	rows, err := o.db.Query(`
		SELECT id, project, status, start_time, 
		       COALESCE(cost, 0), COALESCE(input_tokens, 0), 
		       COALESCE(output_tokens, 0), COALESCE(cache_tokens, 0)
		FROM sessions ORDER BY start_time DESC LIMIT 200
	`)
	if err != nil {
		return fmt.Errorf("opencode query: %w", err)
	}
	defer rows.Close()

	newSessions := make(map[string]*Session)
	for rows.Next() {
		var s Session
		var startTime string
		err := rows.Scan(&s.ID, &s.Project, &s.Status, &startTime,
			&s.Cost, &s.InputTokens, &s.OutputTokens, &s.CacheTokens)
		if err != nil {
			log.Printf("opencode: row scan: %v", err)
			continue
		}
		s.Agent = "OpenCode"
		s.StartTime, _ = time.Parse(time.RFC3339, startTime)
		s.ComputeCacheHitRate()
		s.ComputeDuration()
		newSessions[s.ID] = &s
	}

	o.mu.Lock()
	o.sessions = newSessions
	o.mu.Unlock()

	return rows.Err()
}

func (o *OpenCodeReader) ListSessions() ([]Session, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	sessions := make([]Session, 0, len(o.sessions))
	for _, s := range o.sessions {
		s.ComputeDuration()
		sessions = append(sessions, *s)
	}
	return sessions, nil
}

func (o *OpenCodeReader) GetSession(id string) (*Session, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	s, ok := o.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session %s not found", id)
	}
	s.ComputeDuration()
	return s, nil
}

func (o *OpenCodeReader) GetFileChanges(sessionID string) ([]FileChange, error) {
	rows, err := o.db.Query(`
		SELECT file_path, additions, deletions 
		FROM file_changes WHERE session_id = ?
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("opencode file changes query: %w", err)
	}
	defer rows.Close()

	var changes []FileChange
	for rows.Next() {
		var fc FileChange
		if err := rows.Scan(&fc.Path, &fc.Additions, &fc.Deletions); err != nil {
			continue
		}
		changes = append(changes, fc)
	}
	return changes, rows.Err()
}

func (o *OpenCodeReader) KillSession(id string) error {
	o.mu.RLock()
	s, ok := o.sessions[id]
	o.mu.RUnlock()
	if !ok {
		return fmt.Errorf("session %q not found", id)
	}
	return killProcessByPID(s.PID, id)
}

// Close closes the underlying database connection.
func (o *OpenCodeReader) Close() error {
	if o.db != nil {
		return o.db.Close()
	}
	return nil
}
