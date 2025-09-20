//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd

package arkivformat

import (
	"fmt"
	"os"
)

// getUID returns 0 on Windows as Unix-style UIDs are not available.
func getUID(fi os.FileInfo) int {
	return 0
}

// getGID returns 0 on Windows as Unix-style GIDs are not available.
func getGID(fi os.FileInfo) int {
	return 0
}

// chownBestEffort is a no-op on Windows.
func chownBestEffort(p string, uid, gid int) error {
	return nil
}

// mkfifo is not supported on Windows and returns an explicit error.
func mkfifo(path string, mode uint32) error {
	return fmt.Errorf("FIFO not supported on Windows")
}

