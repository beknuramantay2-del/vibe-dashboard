//go:build !windows

package sources

import (
	"os"
	"syscall"
)

// signalZero returns signal 0 for process existence check (Unix only).
func signalZero() os.Signal {
	return syscall.Signal(0)
}
