package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/vibe-dashboard/vibe-dashboard/rollback"
	"github.com/vibe-dashboard/vibe-dashboard/sources"
	"github.com/vibe-dashboard/vibe-dashboard/store"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// --- DTO types exposed to the frontend ---

type SessionDTO struct {
	ID           string  `json:"id"`
	Agent        string  `json:"agent"`
	Project      string  `json:"project"`
	Status       string  `json:"status"`
	Cost         float64 `json:"cost"`
	InputTokens  int64   `json:"inputTokens"`
	OutputTokens int64   `json:"outputTokens"`
	CacheHitRate float64 `json:"cacheHitRate"`
	Duration     string  `json:"duration"`
	StartTime    string  `json:"startTime"`
	PID          int     `json:"pid"`
}

type FileChangeDTO struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

type SnapshotDTO struct {
	ID        string `json:"id"`
	SessionID string `json:"sessionId"`
	Message   string `json:"message"`
	CreatedAt string `json:"createdAt"`
}

type ResultDTO struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// --- App ---

type App struct {
	ctx      context.Context
	cancel   context.CancelFunc
	readers  []sources.SourceReader
	store    *store.Store
	rollback *rollback.Manager
	logFile  *os.File
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	pollCtx, cancel := context.WithCancel(ctx)
	a.ctx = pollCtx
	a.cancel = cancel

	// Setup logging
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("home dir: %v", err)
		return
	}
	logDir := filepath.Join(home, ".vibe-dashboard")
	if err := os.MkdirAll(logDir, 0700); err != nil {
		log.Printf("log dir: %v", err)
	}
	logPath := filepath.Join(logDir, "vibe-desktop.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err == nil {
		a.logFile = logFile // Store handle — closed in shutdown
		log.SetOutput(logFile)
	}

	// Init store
	a.store, err = store.NewStore()
	if err != nil {
		log.Printf("store init: %v", err)
	}

	// Init rollback manager
	a.rollback, err = rollback.NewManager()
	if err != nil {
		log.Printf("rollback init: %v", err)
	}

	// Connect source readers
	if r, err := sources.NewClaudeReader(); err == nil {
		a.readers = append(a.readers, r)
		log.Println("connected: Claude Code")
	} else {
		log.Printf("claude: %v", err)
	}
	if r, err := sources.NewOpenCodeReader(); err == nil {
		a.readers = append(a.readers, r)
		log.Println("connected: OpenCode")
	} else {
		log.Printf("opencode: %v", err)
	}
	if r, err := sources.NewCodexReader(); err == nil {
		a.readers = append(a.readers, r)
		log.Println("connected: Codex CLI")
	} else {
		log.Printf("codex: %v", err)
	}

	go a.pollLoop()
}

func (a *App) shutdown(ctx context.Context) {
	log.Println("shutting down")

	// Cancel the poll loop
	if a.cancel != nil {
		a.cancel()
	}

	// Close the store
	if a.store != nil {
		if err := a.store.Close(); err != nil {
			log.Printf("store close: %v", err)
		}
	}

	// Close the log file last
	if a.logFile != nil {
		a.logFile.Close()
	}
}

func (a *App) pollLoop() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// Initial refresh
	a.refreshAllSources()

	for {
		select {
		case <-a.ctx.Done():
			log.Println("poll loop stopped")
			return
		case <-ticker.C:
			a.refreshAllSources()
		}
	}
}

func (a *App) refreshAllSources() {
	var allSessions []sources.Session
	for _, reader := range a.readers {
		if err := reader.Refresh(); err != nil {
			log.Printf("refresh %s: %v", reader.Name(), err)
		}
		sessions, err := reader.ListSessions()
		if err != nil {
			log.Printf("list %s: %v", reader.Name(), err)
			continue
		}
		allSessions = append(allSessions, sessions...)

		// Persist to store
		if a.store != nil {
			for _, s := range sessions {
				if err := a.store.SaveSession(s); err != nil {
					log.Printf("save session: %v", err)
				}
			}
		}
	}
	if len(allSessions) > 0 {
		runtime.EventsEmit(a.ctx, "sessions-updated", toSessionDTOs(allSessions))
	}
}

func toSessionDTOs(sessions []sources.Session) []SessionDTO {
	dtos := make([]SessionDTO, len(sessions))
	for i, s := range sessions {
		dtos[i] = SessionDTO{
			ID:           s.ID,
			Agent:        s.Agent,
			Project:      s.Project,
			Status:       s.Status,
			Cost:         s.Cost,
			InputTokens:  s.InputTokens,
			OutputTokens: s.OutputTokens,
			CacheHitRate: s.CacheHitRate,
			Duration:     formatDuration(s.Duration),
			StartTime:    s.StartTime.Format("15:04:05"),
			PID:          s.PID,
		}
	}

	// Sort by start time descending (newest first)
	sort.Slice(dtos, func(i, j int) bool {
		return dtos[i].StartTime > dtos[j].StartTime
	})

	return dtos
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return time.Duration(d).String()
	}
	if m > 0 {
		return time.Duration(time.Duration(m)*time.Minute + time.Duration(s)*time.Second).String()
	}
	return time.Duration(time.Duration(s) * time.Second).String()
}

// --- Frontend bindings ---

// GetSessions returns all sessions from all connected agents.
func (a *App) GetSessions() []SessionDTO {
	var all []sources.Session
	for _, reader := range a.readers {
		sessions, err := reader.ListSessions()
		if err != nil {
			continue
		}
		all = append(all, sessions...)
	}
	return toSessionDTOs(all)
}

// GetFileChanges returns file modifications for a specific session.
func (a *App) GetFileChanges(sessionID string, agent string) []FileChangeDTO {
	for _, reader := range a.readers {
		if reader.Name() == agent || agent == "" {
			changes, err := reader.GetFileChanges(sessionID)
			if err != nil {
				log.Printf("file changes: %v", err)
				return nil
			}
			dtos := make([]FileChangeDTO, len(changes))
			for i, c := range changes {
				dtos[i] = FileChangeDTO{Path: c.Path, Additions: c.Additions, Deletions: c.Deletions}
			}
			return dtos
		}
	}
	return nil
}

// KillSession terminates an active session's process.
func (a *App) KillSession(id string, agent string) ResultDTO {
	for _, reader := range a.readers {
		if reader.Name() == agent || agent == "" {
			if err := reader.KillSession(id); err != nil {
				return ResultDTO{OK: false, Message: err.Error()}
			}
			return ResultDTO{OK: true, Message: "Session terminated"}
		}
	}
	return ResultDTO{OK: false, Message: "Agent not found"}
}

// GetConnectedAgents returns the names of all connected agent sources.
func (a *App) GetConnectedAgents() []string {
	names := make([]string, len(a.readers))
	for i, r := range a.readers {
		names[i] = r.Name()
	}
	return names
}

// CreateSnapshot creates a git stash snapshot for rollback.
func (a *App) CreateSnapshot(sessionID string, repoPath string) ResultDTO {
	if a.rollback == nil {
		return ResultDTO{OK: false, Message: "Rollback manager not initialized"}
	}
	snap, err := a.rollback.CreateSnapshot(sessionID, repoPath)
	if err != nil {
		return ResultDTO{OK: false, Message: err.Error()}
	}
	return ResultDTO{OK: true, Message: "Snapshot created: " + snap.ID}
}

// ListSnapshots returns all saved snapshots.
func (a *App) ListSnapshots() []SnapshotDTO {
	if a.rollback == nil {
		return nil
	}
	snaps, err := a.rollback.ListSnapshots()
	if err != nil {
		log.Printf("list snapshots: %v", err)
		return nil
	}
	dtos := make([]SnapshotDTO, len(snaps))
	for i, s := range snaps {
		dtos[i] = SnapshotDTO{
			ID:        s.ID,
			SessionID: s.SessionID,
			Message:   s.Message,
			CreatedAt: s.CreatedAt.Format(time.RFC3339),
		}
	}
	return dtos
}

// Rollback restores a snapshot.
func (a *App) Rollback(snapshotID string, repoPath string) ResultDTO {
	if a.rollback == nil {
		return ResultDTO{OK: false, Message: "Rollback manager not initialized"}
	}
	if err := a.rollback.Rollback(snapshotID, repoPath); err != nil {
		return ResultDTO{OK: false, Message: err.Error()}
	}
	return ResultDTO{OK: true, Message: "Rollback successful"}
}

// DeleteSnapshot removes a snapshot.
func (a *App) DeleteSnapshot(snapshotID string) ResultDTO {
	if a.rollback == nil {
		return ResultDTO{OK: false, Message: "Rollback manager not initialized"}
	}
	if err := a.rollback.DeleteSnapshot(snapshotID); err != nil {
		return ResultDTO{OK: false, Message: err.Error()}
	}
	return ResultDTO{OK: true, Message: "Snapshot deleted"}
}

// GetAggregatedCost returns the total cost across all sessions.
func (a *App) GetAggregatedCost() float64 {
	if a.store == nil {
		return 0
	}
	total, _ := a.store.GetAggregatedCost(0, "")
	return total
}

// GetCostByAgent returns total cost for a specific agent.
func (a *App) GetCostByAgent(agent string) float64 {
	if a.store == nil {
		return 0
	}
	total, _ := a.store.GetAggregatedCost(0, agent)
	return total
}

// GetCostByHours returns total cost in the last N hours.
func (a *App) GetCostByHours(hours int) float64 {
	if a.store == nil {
		return 0
	}
	total, _ := a.store.GetAggregatedCost(hours, "")
	return total
}
