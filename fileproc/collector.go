// Package fileproc provides functions for collecting and processing files.
package fileproc

// CollectFiles scans the given root directory using the default walker (ProdWalker)
// and returns the collected file paths and the number of files skipped for
// exceeding the size limit or as credential files.
func CollectFiles(root string, opts WalkOptions) ([]string, int, error) {
	w := NewProdWalker(opts)
	files, err := w.Walk(root)

	return files, w.Skipped(), err
}
