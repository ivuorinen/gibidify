// Package fileproc provides functions for file processing.
package fileproc

// FakeWalker implements Walker for testing purposes.
type FakeWalker struct {
	Files []string
	Err   error
}

// Walk returns predetermined file paths or an error, depending on FakeWalker's configuration.
func (fw FakeWalker) Walk(root string) ([]string, error) {
	if fw.Err != nil {
		return nil, fw.Err
	}
	return fw.Files, nil
}
