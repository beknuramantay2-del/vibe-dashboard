package sources

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type ClaudeReader struct {
	baseDir string
	mu      sync.RWMutex
	sessions map[string]*Session
}

type claudeLogEntry struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	SessionID string    `json:"session_id"`
	Project   string    `json:"project"`
	Tokens    *struct {
		Input  int64 `json:"input"`
		Output int64 `json:"output"`
		Cache  int64 `json:"cache_read"`
	} `json:"tokens,omitempty"`
	Cost    float64 `json:"cost,omitempty"`
	Message string  `json:"message,omitempty"`
	Status  string  `json:"status,omitempty"`
	PID     int     `json:"pid,omitempty"`
}

func NewClaudeReader() (*ClaudeReader, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}
	base := filepath.Join(home, ".claude", "projects")
	return &ClaudeReader{
		baseDir:  base,
		sessions: make(map[string]*Session),
	}, nil
}

func (c *ClaudeReader) Name() string { return "Claude Code" }

func (c *ClaudeReader) ListSessions() ([]Session, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sessions := make([]Session, 0, len(c.sessions))
	for _, s := range c.sessions {
		sessions = append(sessions, *s)
	}
	return sessions, nil
}

func (c *ClaudeReader) GetSession(id string) (*Session, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session %s not found", id)
	}
	return s, nil
}

func (c *ClaudeReader) GetFileChanges(sessionID string) ([]FileChange, error) {
	return nil, nil
}

func (c *ClaudeReader) KillSession(id string) error {
	c.mu.RLock()
	s, ok := c.sessions[id]
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("session %s not found", id)
	}
	proc, err := os.FindProcess(s.PID)
	if err != nil {
		return err
	}
	return proc.Signal(os.Interrupt)
}

func (c *ClaudeReader) Watch(callback func(Session)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	filepath.Walk(c.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".jsonl") {
			watcher.Add(path)
			c.parseJSONL(path)
		}
		return nil
	})

	done := make(chan error)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					c.parseJSONL(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Fprintf(os.Stderr, "watcher error: %v\n", err)
			}
		}
	}()
	<-done
	return nil
}

func (c *ClaudeReader) parseJSONL(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry claudeLogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		if entry.SessionID == "" {
			continue
		}

		c.mu.Lock()
		s, exists := c.sessions[entry.SessionID]
		if !exists {
			s = &Session{
				ID:        entry.SessionID,
				Agent:     "claude",
				StartTime: entry.Timestamp,
			}
			c.sessions[entry.SessionID] = s
		}
		if entry.Project != "" {
			s.Project = entry.Project
		}
		if entry.Status != "" {
			s.Status = entry.Status
		}
		if entry.Tokens != nil {
			s.InputTokens += entry.Tokens.Input
			s.OutputTokens += entry.Tokens.Output
			s.CacheTokens += entry.Tokens.Cache
			total := s.InputTokens + s.OutputTokens
			if total > 0 {
				s.CacheHitRate = float64(s.CacheTokens) / float64(total) * 100
			}
		}
		if entry.Cost > 0 {
			s.Cost = entry.Cost
		}
		if entry.PID > 0 {
			s.PID = entry.PID
		}
		s.Duration = time.Since(s.StartTime)
		if entry.Status == "completed" || entry.Status == "terminated" {
			s.Status = "completed"
		}
		c.mu.Unlock()
	}
}
