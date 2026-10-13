//go:build !unix

package fileproc

import (
	"os"
	"path/filepath"
)

// openNoFollow refuses a symlink and opens path for reading. Without
// O_NOFOLLOW the check and the open are separate steps, so a path swapped
// between them is not caught here; validateFile still rejects anything that is
// not a regular file once it is open.
func openNoFollow(path string) (*os.File, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, errSymlink
	}

	return os.Open(filepath.Clean(path))
}
