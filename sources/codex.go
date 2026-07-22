package sources

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type CodexReader struct {
	baseDir  string
	mu       sync.RWMutex
	sessions map[string]*Session
}

type codexLogEntry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Project   string    `json:"project"`
	Input     int64     `json:"input_tokens"`
	Output    int64     `json:"output_tokens"`
	Cost      float64   `json:"cost"`
	Status    string    `json:"status"`
	PID       int       `json:"pid"`
}

func NewCodexReader() (*CodexReader, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}
	base := filepath.Join(home, ".codex", "logs")
	return &CodexReader{
		baseDir:  base,
		sessions: make(map[string]*Session),
	}, nil
}

func (c *CodexReader) Name() string { return "Codex CLI" }

func (c *CodexReader) ListSessions() ([]Session, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	sessions := make([]Session, 0, len(c.sessions))
	for _, s := range c.sessions {
		sessions = append(sessions, *s)
	}
	return sessions, nil
}

func (c *CodexReader) GetSession(id string) (*Session, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session %s not found", id)
	}
	return s, nil
}

func (c *CodexReader) GetFileChanges(sessionID string) ([]FileChange, error) {
	return nil, nil
}

func (c *CodexReader) KillSession(id string) error {
	c.mu.RLock()
	s, ok := c.sessions[id]
	if !ok {
		c.mu.RUnlock()
		return fmt.Errorf("session not found")
	}
	pid := s.PID
	if pid <= 0 {
		c.mu.RUnlock()
		return fmt.Errorf("invalid PID for session")
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		c.mu.RUnlock()
		return fmt.Errorf("process not found")
	}
	c.mu.RUnlock()

	return proc.Signal(os.Interrupt)
}

func (c *CodexReader) Watch(callback func(Session)) error {
	return nil
}

func (c *CodexReader) parseJSONL(path string) {
	fi, err := os.Stat(path)
	if err != nil {
		return
	}
	if fi.Size() > 100*1024*1024 {
		log.Printf("codex: skipping %s (size %d exceeds 100MB limit)", path, fi.Size())
		return
	}

	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	var entries int
	const maxEntries = 100000
	for scanner.Scan() {
		if entries >= maxEntries {
			break
		}
		var entry codexLogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		if entry.ID == "" {
			continue
		}

		c.mu.Lock()
		s, exists := c.sessions[entry.ID]
		if !exists {
			s = &Session{
				ID:        entry.ID,
				Agent:     "codex",
				StartTime: entry.Timestamp,
			}
			c.sessions[entry.ID] = s
		}
		if entry.Project != "" {
			s.Project = entry.Project
		}
		if entry.Status != "" {
			if entry.Status == "completed" || entry.Status == "terminated" {
				s.Status = "completed"
			} else {
				s.Status = "active"
			}
		}
		s.InputTokens += entry.Input
		s.OutputTokens += entry.Output
		if entry.Cost > 0 {
			s.Cost = entry.Cost
		}
		if entry.PID > 0 {
			s.PID = entry.PID
		}
		s.Duration = time.Since(s.StartTime)
		c.mu.Unlock()
		entries++
	}

	if err := scanner.Err(); err != nil {
		log.Printf("codex: scanner error reading %s: %v", path, err)
	}
}
