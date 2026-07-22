package sources

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// CodexReader reads Codex CLI JSONL session logs.
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
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return nil, fmt.Errorf("codex logs dir not found: %s", base)
	}
	return &CodexReader{
		baseDir:  base,
		sessions: make(map[string]*Session),
	}, nil
}

func (c *CodexReader) Name() string { return "Codex CLI" }

// Refresh scans all JSONL files under the Codex logs directory.
func (c *CodexReader) Refresh() error {
	files, err := filepath.Glob(filepath.Join(c.baseDir, "**", "*.jsonl"))
	if err != nil {
		return fmt.Errorf("glob codex logs: %w", err)
	}
	deeper, _ := filepath.Glob(filepath.Join(c.baseDir, "*.jsonl"))
	files = append(files, deeper...)

	seen := make(map[string]bool, len(files))
	for _, f := range files {
		if !seen[f] {
			seen[f] = true
			c.parseJSONL(f)
		}
	}
	return nil
}

func (c *CodexReader) ListSessions() ([]Session, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	sessions := make([]Session, 0, len(c.sessions))
	for _, s := range c.sessions {
		s.ComputeDuration()
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
	s.ComputeDuration()
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
		return fmt.Errorf("session %q not found", id)
	}
	pid := s.PID
	c.mu.RUnlock()

	if pid <= 0 {
		return fmt.Errorf("no valid PID for session %q", id)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process %d not found: %w", pid, err)
	}

	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return fmt.Errorf("process %d not running: %w", pid, err)
	}

	return proc.Signal(os.Interrupt)
}

func (c *CodexReader) parseJSONL(path string) {
	fi, err := os.Stat(path)
	if err != nil {
		return
	}
	if fi.Size() > maxFileSize {
		log.Printf("codex: skipping %s (%.1f MB exceeds limit)", path, float64(fi.Size())/(1024*1024))
		return
	}

	f, err := os.Open(path)
	if err != nil {
		log.Printf("codex: cannot open %s: %v", path, err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, scanBufferSize), scanBufferSize)

	var entries int
	for scanner.Scan() {
		if entries >= maxLogEntries {
			log.Printf("codex: reached %d entry limit in %s", maxLogEntries, path)
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
				Agent:     "Codex CLI",
				Status:    "active",
				StartTime: entry.Timestamp,
			}
			c.sessions[entry.ID] = s
		}
		if entry.Project != "" {
			s.Project = entry.Project
		}
		s.InputTokens += entry.Input
		s.OutputTokens += entry.Output
		s.ComputeCacheHitRate()
		if entry.Cost > 0 {
			s.Cost += entry.Cost
		}
		if entry.PID > 0 {
			s.PID = entry.PID
		}
		if entry.Status == "completed" || entry.Status == "terminated" {
			s.Status = "completed"
			s.EndTime = entry.Timestamp
		} else if entry.Status != "" {
			s.Status = "active"
		}
		c.mu.Unlock()
		entries++
	}

	if err := scanner.Err(); err != nil {
		log.Printf("codex: scanner error reading %s: %v", path, err)
	}
}
