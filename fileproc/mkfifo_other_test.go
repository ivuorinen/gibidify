//go:build !unix

package fileproc_test

import "errors"

// mkfifo reports that named pipes are unavailable on this platform.
func mkfifo(string) error {
	return errors.New("named pipes are not supported on this platform")
}
