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

// Snapshot represents a saved state before an AI session's changes.
type Snapshot struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	Files     []string  `json:"files"`
}

// Manager handles creating and restoring rollback snapshots via git stash.
type Manager struct {
	snapshotsDir string
}

// NewManager creates a new rollback manager with storage at ~/.vibe-dashboard/snapshots.
func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(home, ".vibe-dashboard", "snapshots")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	return &Manager{snapshotsDir: dir}, nil
}

func sanitizeSessionID(id string) string {
	h := sha256.Sum256([]byte(id))
	return fmt.Sprintf("%x", h[:12])
}

// validateRepoPath ensures the path is safe: absolute, exists, is a directory, and is a git repo.
func validateRepoPath(repoPath string) (string, error) {
	cleanPath, err := filepath.Abs(repoPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Resolve symlinks to prevent traversal
	cleanPath, err = filepath.EvalSymlinks(cleanPath)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path: %w", err)
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		return "", fmt.Errorf("path does not exist: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", cleanPath)
	}

	if !isGitRepo(cleanPath) {
		return "", fmt.Errorf("not a git repository: %s", cleanPath)
	}

	return cleanPath, nil
}

// CreateSnapshot saves the current state of a repo via git stash.
func (m *Manager) CreateSnapshot(sessionID string, repoPath string) (*Snapshot, error) {
	cleanPath, err := validateRepoPath(repoPath)
	if err != nil {
		return nil, err
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
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git stash: %w (output: %s)", err, strings.TrimSpace(string(output)))
	}

	snapshot.Message = msg

	// Write snapshot metadata
	data := fmt.Sprintf("ID=%s\nSession=%s\nTime=%s\nMessage=%s\n",
		snapshot.ID, sessionTag, snapshot.CreatedAt.Format(time.RFC3339), msg)

	snapPath := filepath.Join(m.snapshotsDir, snapshot.ID+".snap")
	if err := os.WriteFile(snapPath, []byte(data), 0600); err != nil {
		return nil, fmt.Errorf("save snapshot metadata: %w", err)
	}

	return snapshot, nil
}

// ListSnapshots returns all saved snapshots.
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

// parseSnapshot reads key=value formatted snapshot metadata.
// Uses "=" as delimiter instead of ":" to avoid conflict with RFC3339 timestamps.
func parseSnapshot(data string) *Snapshot {
	lines := strings.Split(strings.TrimSpace(data), "\n")
	if len(lines) < 3 {
		return nil
	}
	s := &Snapshot{}
	for _, line := range lines {
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			// Fallback: try ":" delimiter for old-format snapshots
			idx = strings.IndexByte(line, ':')
			if idx < 0 {
				continue
			}
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			switch key {
			case "ID":
				s.ID = val
			case "Session":
				s.SessionID = val
			case "Time":
				s.CreatedAt, _ = time.Parse(time.RFC3339, val)
			case "Message":
				s.Message = val
			}
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		switch key {
		case "ID":
			s.ID = val
		case "Session":
			s.SessionID = val
		case "Time":
			s.CreatedAt, _ = time.Parse(time.RFC3339, val)
		case "Message":
			s.Message = val
		}
	}
	if s.ID == "" {
		return nil
	}
	return s
}

// Rollback restores a snapshot by applying the matching git stash.
func (m *Manager) Rollback(snapshotID string, repoPath string) error {
	cleanPath, err := validateRepoPath(repoPath)
	if err != nil {
		return err
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
			colonIdx := strings.IndexByte(line, ':')
			if colonIdx < 0 {
				continue
			}
			stashRef := strings.TrimSpace(line[:colonIdx])

			applyCtx, applyCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer applyCancel()
			applyCmd := exec.CommandContext(applyCtx, "git", "stash", "apply", stashRef)
			applyCmd.Dir = cleanPath
			output, err := applyCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("git stash apply %s: %w (output: %s)", stashRef, err, strings.TrimSpace(string(output)))
			}
			return nil
		}
	}

	return fmt.Errorf("snapshot %q not found in git stash list", snapshotID)
}

// DeleteSnapshot removes a snapshot metadata file.
func (m *Manager) DeleteSnapshot(snapshotID string) error {
	// Sanitize to prevent path traversal
	if strings.Contains(snapshotID, "/") || strings.Contains(snapshotID, "..") {
		return fmt.Errorf("invalid snapshot ID")
	}
	snapPath := filepath.Join(m.snapshotsDir, snapshotID+".snap")
	return os.Remove(snapPath)
}

func isGitRepo(dir string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}
