package sources

import "time"

// Session represents a unified AI coding agent session.
type Session struct {
	ID           string        `json:"id"`
	Agent        string        `json:"agent"`
	Project      string        `json:"project"`
	Status       string        `json:"status"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time,omitempty"`
	Duration     time.Duration `json:"duration"`
	Cost         float64       `json:"cost"`
	InputTokens  int64         `json:"input_tokens"`
	OutputTokens int64         `json:"output_tokens"`
	CacheTokens  int64         `json:"cache_tokens"`
	CacheHitRate float64       `json:"cache_hit_rate"`
	PID          int           `json:"pid,omitempty"`
}

// ComputeCacheHitRate recalculates the cache hit rate from token counts.
func (s *Session) ComputeCacheHitRate() {
	total := s.InputTokens + s.CacheTokens
	if total > 0 {
		s.CacheHitRate = float64(s.CacheTokens) / float64(total) * 100
	}
}

// ComputeDuration sets the duration based on start and end time.
// For active sessions it uses the current time.
func (s *Session) ComputeDuration() {
	if s.Status == "completed" || s.Status == "terminated" {
		if !s.EndTime.IsZero() {
			s.Duration = s.EndTime.Sub(s.StartTime)
			return
		}
	}
	s.Duration = time.Since(s.StartTime)
}

// FileChange represents a file modified during a session.
type FileChange struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Diff      string `json:"diff,omitempty"`
}

// SourceReader is the interface all agent data sources must implement.
type SourceReader interface {
	// Name returns the human-readable name of this agent source.
	Name() string

	// Refresh re-reads data from the underlying source.
	// Must be called before ListSessions returns up-to-date data.
	Refresh() error

	// ListSessions returns all known sessions.
	ListSessions() ([]Session, error)

	// GetSession returns a single session by ID.
	GetSession(id string) (*Session, error)

	// GetFileChanges returns files modified in the given session.
	GetFileChanges(sessionID string) ([]FileChange, error)

	// KillSession sends an interrupt signal to the session's process.
	KillSession(id string) error
}
