//go:build windows

package sources

import "os"

// signalZero is a no-op on Windows (not supported).
// The caller skips this check on Windows.
func signalZero() os.Signal {
	return os.Kill
}
