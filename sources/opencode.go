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
	}

	var db *sql.DB
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			db, err = sql.Open("sqlite", p)
			if err != nil {
				continue
			}
			if err = db.Ping(); err == nil {
				break
			}
			db.Close()
		}
	}

	if db == nil {
		return nil, fmt.Errorf("opencode database not found in any known path")
	}

	r := &OpenCodeReader{
		db:       db,
		sessions: make(map[string]*Session),
	}
	r.loadSessions()
	return r, nil
}

func (o *OpenCodeReader) Name() string { return "OpenCode" }

func (o *OpenCodeReader) loadSessions() {
	o.mu.Lock()
	defer o.mu.Unlock()

	rows, err := o.db.Query(`
		SELECT id, project, status, start_time, 
			   COALESCE(cost, 0), COALESCE(input_tokens, 0), 
			   COALESCE(output_tokens, 0), COALESCE(cache_tokens, 0)
		FROM sessions ORDER BY start_time DESC LIMIT 100
	`)
	if err != nil {
		log.Printf("opencode: query failed: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var s Session
		var startTime string
		err := rows.Scan(&s.ID, &s.Project, &s.Status, &startTime,
			&s.Cost, &s.InputTokens, &s.OutputTokens, &s.CacheTokens)
		if err != nil {
			log.Printf("opencode: row scan failed: %v", err)
			continue
		}
		s.Agent = "opencode"
		s.StartTime, _ = time.Parse(time.RFC3339, startTime)
		s.Duration = time.Since(s.StartTime)
		total := s.InputTokens + s.OutputTokens
		if total > 0 {
			s.CacheHitRate = float64(s.CacheTokens) / float64(total) * 100
		}
		o.sessions[s.ID] = &s
	}
}

func (o *OpenCodeReader) ListSessions() ([]Session, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	sessions := make([]Session, 0, len(o.sessions))
	for _, s := range o.sessions {
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
	return s, nil
}

func (o *OpenCodeReader) GetFileChanges(sessionID string) ([]FileChange, error) {
	rows, err := o.db.Query(`
		SELECT file_path, additions, deletions 
		FROM file_changes WHERE session_id = ?
	`, sessionID)
	if err != nil {
		return nil, err
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
	return changes, nil
}

func (o *OpenCodeReader) Watch(callback func(Session)) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		o.mu.RLock()
		oldCount := len(o.sessions)
		o.mu.RUnlock()
		o.loadSessions()
		o.mu.RLock()
		newCount := len(o.sessions)
		o.mu.RUnlock()
		if newCount > oldCount {
			o.mu.RLock()
			for _, s := range o.sessions {
				callback(*s)
			}
			o.mu.RUnlock()
		}
	}
	return nil
}

func (o *OpenCodeReader) KillSession(id string) error {
	_, err := o.db.Exec("UPDATE sessions SET status = 'terminated' WHERE id = ?", id)
	return err
}
