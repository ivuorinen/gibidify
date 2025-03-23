// Package fileproc provides a writer for the output of the file processor.
//
// The StartWriter function writes the output in the specified format.
// The formatMarkdown function formats the output in Markdown format.
// The detectLanguage function tries to infer the code block language from the file extension.
// The OutputData struct represents the full output structure.
// The FileData struct represents a single file's path and content.
package fileproc

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// FileData represents a single file's path and content.
type FileData struct {
	Path    string `json:"path" yaml:"path"`
	Content string `json:"content" yaml:"content"`
}

// OutputData represents the full output structure.
type OutputData struct {
	Prefix string     `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Files  []FileData `json:"files" yaml:"files"`
	Suffix string     `json:"suffix,omitempty" yaml:"suffix,omitempty"`
}

// StartWriter writes the output in the specified format.
func StartWriter(outFile *os.File, writeCh <-chan WriteRequest, done chan<- struct{}, format string, prefix, suffix string) {
	var files []FileData

	// Read from channel until closed
	for req := range writeCh {
		files = append(files, FileData(req))
	}

	// Create output struct
	output := OutputData{Prefix: prefix, Files: files, Suffix: suffix}

	// Serialize based on format
	var outputData []byte
	var err error

	switch format {
	case "json":
		outputData, err = json.MarshalIndent(output, "", "  ")
	case "yaml":
		outputData, err = yaml.Marshal(output)
	case "markdown":
		outputData = []byte(formatMarkdown(output))
	default:
		err = fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		logrus.Errorf("Error encoding output: %v", err)
		close(done)
		return
	}

	// Write to file
	if _, err := outFile.Write(outputData); err != nil {
		logrus.Errorf("Error writing to file: %v", err)
	}

	close(done)
}

func formatMarkdown(output OutputData) string {
	markdown := "# " + output.Prefix + "\n\n"

	for _, file := range output.Files {
		markdown += fmt.Sprintf("## File: `%s`\n```%s\n%s\n```\n\n", file.Path, detectLanguage(file.Path), file.Content)
	}

	markdown += "# " + output.Suffix
	return markdown
}

// detectLanguage tries to infer code block language from file extension.
func detectLanguage(filename string) string {
	if len(filename) < 3 {
		return ""
	}
	switch {
	case len(filename) >= 3 && filename[len(filename)-3:] == ".go":
		return "go"
	case len(filename) >= 3 && filename[len(filename)-3:] == ".py":
		return "python"
	case len(filename) >= 2 && filename[len(filename)-2:] == ".c":
		return "c"
	case len(filename) >= 3 && filename[len(filename)-3:] == ".js":
		return "javascript"
	default:
		return ""
	}
}
