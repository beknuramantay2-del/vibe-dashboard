package sources

import "time"

type Session struct {
	ID        string    `json:"id"`
	Agent     string    `json:"agent"`
	Project   string    `json:"project"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	Duration  time.Duration `json:"duration"`
	Cost      float64   `json:"cost"`
	InputTokens  int64  `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
	CacheTokens  int64  `json:"cache_tokens"`
	CacheHitRate float64 `json:"cache_hit_rate"`
	PID       int       `json:"pid,omitempty"`
}

type FileChange struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Diff      string `json:"diff,omitempty"`
}

type SourceReader interface {
	Name() string
	ListSessions() ([]Session, error)
	GetSession(id string) (*Session, error)
	GetFileChanges(sessionID string) ([]FileChange, error)
	Watch(callback func(Session)) error
	KillSession(id string) error
}
