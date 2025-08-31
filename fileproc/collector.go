// Package fileproc provides functions for collecting and processing files.
package fileproc

// CollectFiles scans the given root directory using the default walker (ProdWalker)
// and returns a slice of file paths.
func CollectFiles(root string) ([]string, error) {
	w := NewProdWalker()

	return w.Walk(root)
}
