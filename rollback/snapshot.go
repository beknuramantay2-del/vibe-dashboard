package rollback

import (
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

func (m *Manager) CreateSnapshot(sessionID string, repoPath string) (*Snapshot, error) {
	if !isGitRepo(repoPath) {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	snapshot := &Snapshot{
		ID:        fmt.Sprintf("snap_%d", time.Now().UnixNano()),
		SessionID: sessionID,
		CreatedAt: time.Now(),
	}

	sanitizedID := strings.ReplaceAll(sessionID, "\n", " ")
	sanitizedID = strings.ReplaceAll(sanitizedID, "\r", " ")
	sanitizedID = strings.ReplaceAll(sanitizedID, "\"", "'")
	msg := fmt.Sprintf("vibe-dashboard: snapshot before session %s", sanitizedID)
	cmd := exec.Command("git", "stash", "push", "-m", msg)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git stash: %w", err)
	}

	snapshot.Message = msg

	data := fmt.Sprintf("ID:%s\nSession:%s\nTime:%s\nMessage:%s\n",
		snapshot.ID, snapshot.SessionID, snapshot.CreatedAt.Format(time.RFC3339), msg)

	snapPath := filepath.Join(m.snapshotsDir, snapshot.ID+".snap")
	if _, err := os.Lstat(snapPath); err == nil {
		return nil, fmt.Errorf("snapshot path already exists: %s", snapPath)
	}
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
	if !isGitRepo(repoPath) {
		return fmt.Errorf("not a git repository: %s", repoPath)
	}

	cmd := exec.Command("git", "stash", "list")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git stash list: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, snapshotID) || strings.Contains(line, "vibe-dashboard: snapshot") {
			stashRef := strings.Split(line, ":")[0]
			if !strings.HasPrefix(stashRef, "stash@{") {
				continue
			}
			applyCmd := exec.Command("git", "stash", "apply", stashRef)
			applyCmd.Dir = repoPath
			if err := applyCmd.Run(); err != nil {
				return fmt.Errorf("git stash apply %s: %w", stashRef, err)
			}
			return nil
		}
	}

	return fmt.Errorf("snapshot %s not found in git stash", snapshotID)
}

func isGitRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	return cmd.Run() == nil
}
