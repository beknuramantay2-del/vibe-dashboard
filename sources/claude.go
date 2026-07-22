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

// ClaudeReader reads Claude Code JSONL session logs.
type ClaudeReader struct {
	baseDir  string
	mu       sync.RWMutex
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

const (
	maxFileSize    = 100 * 1024 * 1024 // 100 MB
	maxLogEntries  = 100_000
	scanBufferSize = 1024 * 1024
)

func NewClaudeReader() (*ClaudeReader, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}
	base := filepath.Join(home, ".claude", "projects")
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return nil, fmt.Errorf("claude projects dir not found: %s", base)
	}
	return &ClaudeReader{
		baseDir:  base,
		sessions: make(map[string]*Session),
	}, nil
}

func (c *ClaudeReader) Name() string { return "Claude Code" }

// Refresh scans all JSONL files under the Claude projects directory.
func (c *ClaudeReader) Refresh() error {
	files, err := filepath.Glob(filepath.Join(c.baseDir, "**", "*.jsonl"))
	if err != nil {
		return fmt.Errorf("glob claude logs: %w", err)
	}
	// Also try one level deeper
	deeper, _ := filepath.Glob(filepath.Join(c.baseDir, "**", "**", "*.jsonl"))
	files = append(files, deeper...)

	// Deduplicate
	seen := make(map[string]bool, len(files))
	for _, f := range files {
		if !seen[f] {
			seen[f] = true
			c.parseJSONL(f)
		}
	}
	return nil
}

func (c *ClaudeReader) ListSessions() ([]Session, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sessions := make([]Session, 0, len(c.sessions))
	for _, s := range c.sessions {
		s.ComputeDuration()
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
	s.ComputeDuration()
	return s, nil
}

func (c *ClaudeReader) GetFileChanges(sessionID string) ([]FileChange, error) {
	// Claude Code doesn't expose file changes in its JSONL logs
	return nil, nil
}

func (c *ClaudeReader) KillSession(id string) error {
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

	// Verify the process is still running before sending signal
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return fmt.Errorf("process %d not running: %w", pid, err)
	}

	return proc.Signal(os.Interrupt)
}

func (c *ClaudeReader) parseJSONL(path string) {
	fi, err := os.Stat(path)
	if err != nil {
		return
	}
	if fi.Size() > maxFileSize {
		log.Printf("claude: skipping %s (%.1f MB exceeds limit)", path, float64(fi.Size())/(1024*1024))
		return
	}

	f, err := os.Open(path)
	if err != nil {
		log.Printf("claude: cannot open %s: %v", path, err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, scanBufferSize), scanBufferSize)

	var entries int
	for scanner.Scan() {
		if entries >= maxLogEntries {
			log.Printf("claude: reached %d entry limit in %s", maxLogEntries, path)
			break
		}

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
				Agent:     "Claude Code",
				Status:    "active",
				StartTime: entry.Timestamp,
			}
			c.sessions[entry.SessionID] = s
		}
		if entry.Project != "" {
			s.Project = entry.Project
		}
		if entry.Tokens != nil {
			s.InputTokens += entry.Tokens.Input
			s.OutputTokens += entry.Tokens.Output
			s.CacheTokens += entry.Tokens.Cache
			s.ComputeCacheHitRate()
		}
		if entry.Cost > 0 {
			s.Cost += entry.Cost
		}
		if entry.PID > 0 {
			s.PID = entry.PID
		}
		if entry.Status == "completed" || entry.Status == "terminated" {
			s.Status = "completed"
			s.EndTime = entry.Timestamp
		}
		c.mu.Unlock()
		entries++
	}

	if err := scanner.Err(); err != nil {
		log.Printf("claude: scanner error reading %s: %v", path, err)
	}
}
