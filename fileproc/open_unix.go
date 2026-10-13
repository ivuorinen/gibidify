//go:build unix

package fileproc

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
)

// openNoFollow opens path for reading without following a final symlink and
// without blocking on a FIFO. The kernel refuses a symlink with ELOOP (EMLINK on
// FreeBSD), so the check and the open are one step and cannot be raced.
func openNoFollow(path string) (*os.File, error) {
	f, err := os.OpenFile(filepath.Clean(path), os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if errors.Is(err, syscall.ELOOP) || errors.Is(err, syscall.EMLINK) {
		return nil, errSymlink
	}

	return f, err
}
