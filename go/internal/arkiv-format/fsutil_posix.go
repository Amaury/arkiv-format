//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package arkivformat

import (
	"os"
	"syscall"
)

// getUID extracts the UID from FileInfo on Unix platforms.
func getUID(fi os.FileInfo) int {
	st := fi.Sys().(*syscall.Stat_t)
	return int(st.Uid)
}

// getGID extracts the GID from FileInfo on Unix platforms.
func getGID(fi os.FileInfo) int {
	st := fi.Sys().(*syscall.Stat_t)
	return int(st.Gid)
}

// chownBestEffort attempts to change ownership; ignore errors when not permitted.
func chownBestEffort(p string, uid, gid int) error {
	return os.Lchown(p, uid, gid)
}

// mkfifo creates a named pipe using mknod on Unix.
func mkfifo(path string, mode uint32) error {
	return syscall.Mkfifo(path, mode)
}

