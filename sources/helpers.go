package sources

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// findJSONLFiles recursively finds all .jsonl files under a directory.
// Go's filepath.Glob does NOT support "**" (recursive glob), so we use WalkDir.
func findJSONLFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip directories we can't read
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".jsonl") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// killProcessByPID sends an interrupt signal to a process.
// Cross-platform: works on Windows (os.Kill) and Unix (os.Interrupt).
func killProcessByPID(pid int, sessionID string) error {
	if pid <= 0 {
		return fmt.Errorf("no valid PID for session %q", sessionID)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process %d not found: %w", pid, err)
	}

	// On Unix, check if process is running first.
	// On Windows, FindProcess always succeeds — we just try to signal directly.
	if runtime.GOOS != "windows" {
		// Signal 0 checks existence without sending a signal (Unix only)
		if err := proc.Signal(signalZero()); err != nil {
			return fmt.Errorf("process %d not running: %w", pid, err)
		}
	}

	// Use Interrupt on Unix, Kill on Windows (Windows doesn't support Interrupt)
	sig := os.Interrupt
	if runtime.GOOS == "windows" {
		sig = os.Kill
	}
	return proc.Signal(sig)
}
