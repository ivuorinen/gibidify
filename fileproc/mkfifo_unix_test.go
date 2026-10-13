//go:build unix

package fileproc_test

import "syscall"

// mkfifo creates a named pipe for walker tests.
func mkfifo(path string) error {
	return syscall.Mkfifo(path, 0o600)
}
