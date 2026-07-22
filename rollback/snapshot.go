package rollback

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Snapshot struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	Files     []string  `json:"files"`
}

type Manager struct {
	snapshotsDir string
}

func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(home, ".vibe-dashboard", "snapshots")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return &Manager{snapshotsDir: dir}, nil
}

func sanitizeSessionID(id string) string {
	h := sha256.Sum256([]byte(id))
	return fmt.Sprintf("%x", h[:12])
}

func (m *Manager) CreateSnapshot(sessionID string, repoPath string) (*Snapshot, error) {
	cleanPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	if !isGitRepo(cleanPath) {
		return nil, fmt.Errorf("not a git repository: %s", cleanPath)
	}

	sessionTag := sanitizeSessionID(sessionID)

	snapshot := &Snapshot{
		ID:        fmt.Sprintf("snap_%d", time.Now().UnixNano()),
		SessionID: sessionID,
		CreatedAt: time.Now(),
	}

	msg := fmt.Sprintf("vibe-dashboard: snapshot before session %s", sessionTag)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "stash", "push", "-m", msg)
	cmd.Dir = cleanPath
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git stash: %w", err)
	}

	snapshot.Message = msg

	data := fmt.Sprintf("ID:%s\nSession:%s\nTime:%s\nMessage:%s\n",
		snapshot.ID, sessionTag, snapshot.CreatedAt.Format(time.RFC3339), msg)

	snapPath := filepath.Join(m.snapshotsDir, snapshot.ID+".snap")
	if err := os.WriteFile(snapPath, []byte(data), 0600); err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (m *Manager) ListSnapshots() ([]Snapshot, error) {
	entries, err := os.ReadDir(m.snapshotsDir)
	if err != nil {
		return nil, err
	}

	var snapshots []Snapshot
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".snap") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(m.snapshotsDir, e.Name()))
		if err != nil {
			continue
		}
		s := parseSnapshot(string(data))
		if s != nil {
			snapshots = append(snapshots, *s)
		}
	}
	return snapshots, nil
}

func parseSnapshot(data string) *Snapshot {
	lines := strings.Split(strings.TrimSpace(data), "\n")
	if len(lines) < 3 {
		return nil
	}
	s := &Snapshot{}
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		switch parts[0] {
		case "ID":
			s.ID = strings.TrimSpace(parts[1])
		case "Session":
			s.SessionID = strings.TrimSpace(parts[1])
		case "Time":
			s.CreatedAt, _ = time.Parse(time.RFC3339, strings.TrimSpace(parts[1]))
		case "Message":
			s.Message = strings.TrimSpace(parts[1])
		}
	}
	return s
}

func (m *Manager) Rollback(snapshotID string, repoPath string) error {
	cleanPath, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	if !isGitRepo(cleanPath) {
		return fmt.Errorf("not a git repository: %s", cleanPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	listCmd := exec.CommandContext(ctx, "git", "stash", "list")
	listCmd.Dir = cleanPath
	out, err := listCmd.Output()
	if err != nil {
		return fmt.Errorf("git stash list: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, snapshotID) || strings.Contains(line, "vibe-dashboard: snapshot") {
			parts := strings.SplitN(line, ":", 2)
			stashRef := strings.TrimSpace(parts[0])

			applyCtx, applyCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer applyCancel()
			applyCmd := exec.CommandContext(applyCtx, "git", "stash", "apply", stashRef)
			applyCmd.Dir = cleanPath
			if err := applyCmd.Run(); err != nil {
				return fmt.Errorf("git stash apply: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("snapshot not found")
}

func isGitRepo(dir string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	return cmd.Run() == nil
}
