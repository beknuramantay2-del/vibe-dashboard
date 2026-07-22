package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/vibe-dashboard/vibe-dashboard/rollback"
	"github.com/vibe-dashboard/vibe-dashboard/sources"
	"github.com/vibe-dashboard/vibe-dashboard/store"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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

type App struct {
	ctx      context.Context
	readers  []sources.SourceReader
	store    *store.Store
	rollback *rollback.Manager
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	home, _ := os.UserHomeDir()
	logDir := filepath.Join(home, ".vibe-dashboard")
	os.MkdirAll(logDir, 0700)
	logPath := filepath.Join(logDir, "vibe-desktop.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err == nil {
		log.SetOutput(logFile)
	}

	a.store, err = store.NewStore()
	if err != nil {
		log.Printf("store init: %v", err)
	}

	a.rollback, err = rollback.NewManager()
	if err != nil {
		log.Printf("rollback init: %v", err)
	}

	if r, err := sources.NewClaudeReader(); err == nil {
		a.readers = append(a.readers, r)
		log.Println("connected to Claude Code")
	}
	if r, err := sources.NewOpenCodeReader(); err == nil {
		a.readers = append(a.readers, r)
		log.Println("connected to OpenCode")
	}
	if r, err := sources.NewCodexReader(); err == nil {
		a.readers = append(a.readers, r)
		log.Println("connected to Codex CLI")
	}

	go a.pollLoop()
}

func (a *App) shutdown(ctx context.Context) {
	log.Println("shutting down")
}

func (a *App) pollLoop() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var allSessions []sources.Session
		for _, reader := range a.readers {
			sessions, err := reader.ListSessions()
			if err != nil {
				log.Printf("error reading %s: %v", reader.Name(), err)
				continue
			}
			allSessions = append(allSessions, sessions...)
			if a.store != nil {
				for _, s := range sessions {
					if err := a.store.SaveSession(s); err != nil {
						log.Printf("error saving: %v", err)
					}
				}
			}
		}
		if len(allSessions) > 0 {
			runtime.EventsEmit(a.ctx, "sessions-updated", toSessionDTOs(allSessions))
		}
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
			Duration:     s.Duration.Round(time.Second).String(),
			StartTime:    s.StartTime.Format("15:04:05"),
			PID:          s.PID,
		}
	}
	return dtos
}

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

func (a *App) GetFileChanges(sessionID string, agent string) []FileChangeDTO {
	for _, reader := range a.readers {
		if reader.Name() == agent || agent == "" {
			changes, err := reader.GetFileChanges(sessionID)
			if err != nil {
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

func (a *App) KillSession(id string, agent string) string {
	for _, reader := range a.readers {
		if reader.Name() == agent || agent == "" {
			if err := reader.KillSession(id); err != nil {
				return err.Error()
			}
			return "ok"
		}
	}
	return "reader not found"
}

func (a *App) GetConnectedAgents() []string {
	names := make([]string, len(a.readers))
	for i, r := range a.readers {
		names[i] = r.Name()
	}
	return names
}

func (a *App) CreateSnapshot(sessionID string, repoPath string) *SnapshotDTO {
	snap, err := a.rollback.CreateSnapshot(sessionID, repoPath)
	if err != nil {
		return nil
	}
	return &SnapshotDTO{
		ID:        snap.ID,
		SessionID: snap.SessionID,
		Message:   snap.Message,
		CreatedAt: snap.CreatedAt.Format(time.RFC3339),
	}
}

func (a *App) ListSnapshots() []SnapshotDTO {
	snaps, err := a.rollback.ListSnapshots()
	if err != nil {
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

func (a *App) Rollback(snapshotID string, repoPath string) string {
	if err := a.rollback.Rollback(snapshotID, repoPath); err != nil {
		return err.Error()
	}
	return "ok"
}

func (a *App) GetAggregatedCost() float64 {
	if a.store == nil {
		return 0
	}
	total, _ := a.store.GetAggregatedCost(0, "")
	return total
}
