package fileproc

// FileData represents a single file's path and content.
type FileData struct {
	Path     string `json:"path"     yaml:"path"`
	Content  string `json:"content"  yaml:"content"`
	Language string `json:"language" yaml:"language"`
}

// OutputData represents the full output structure.
type OutputData struct {
	Prefix string     `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Suffix string     `json:"suffix,omitempty" yaml:"suffix,omitempty"`
	Files  []FileData `json:"files"            yaml:"files"`
}

// FormatWriter defines the interface for format-specific writers.
type FormatWriter interface {
	Start(prefix, suffix string) error
	WriteFile(req WriteRequest) error
	Close() error
}

// detectLanguage tries to infer the code block language from the file extension.
func detectLanguage(filePath string) string {
	registry := GetDefaultRegistry()
	return registry.GetLanguage(filePath)
}
