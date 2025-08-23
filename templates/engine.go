package templates

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Engine handles template processing and output generation.
type Engine struct {
	template OutputTemplate
	context  TemplateContext
}

// NewEngine creates a new template engine with the specified template.
func NewEngine(templateName string, context TemplateContext) (*Engine, error) {
	tmpl, exists := BuiltinTemplates[templateName]
	if !exists {
		return nil, fmt.Errorf("template '%s' not found", templateName)
	}

	// Apply custom variables to context
	if context.Variables == nil {
		context.Variables = make(map[string]string)
	}

	// Merge template variables with context variables
	for k, v := range tmpl.Variables {
		if _, exists := context.Variables[k]; !exists {
			context.Variables[k] = v
		}
	}

	return &Engine{
		template: tmpl,
		context:  context,
	}, nil
}

// NewEngineWithCustomTemplate creates a new template engine with a custom template.
func NewEngineWithCustomTemplate(customTemplate OutputTemplate, context TemplateContext) *Engine {
	if context.Variables == nil {
		context.Variables = make(map[string]string)
	}

	// Merge template variables with context variables
	for k, v := range customTemplate.Variables {
		if _, exists := context.Variables[k]; !exists {
			context.Variables[k] = v
		}
	}

	return &Engine{
		template: customTemplate,
		context:  context,
	}
}

// RenderHeader renders the template header.
func (e *Engine) RenderHeader() (string, error) {
	return e.renderTemplate(e.template.Header, e.context)
}

// RenderFooter renders the template footer.
func (e *Engine) RenderFooter() (string, error) {
	return e.renderTemplate(e.template.Footer, e.context)
}

// RenderFileHeader renders the file header for a specific file.
func (e *Engine) RenderFileHeader(fileCtx FileContext) (string, error) {
	return e.renderTemplate(e.template.FileHeader, fileCtx)
}

// RenderFileFooter renders the file footer for a specific file.
func (e *Engine) RenderFileFooter(fileCtx FileContext) (string, error) {
	return e.renderTemplate(e.template.FileFooter, fileCtx)
}

// RenderFileContent renders file content according to template options.
func (e *Engine) RenderFileContent(fileCtx FileContext) (string, error) {
	content := fileCtx.Content

	// Apply markdown-specific formatting if needed
	if e.template.Format == "markdown" && e.template.Markdown.UseCodeBlocks {
		// Content is wrapped in code blocks via FileHeader/FileFooter
		return content, nil
	}

	// Apply line length limits if specified
	if e.template.Markdown.MaxLineLength > 0 {
		content = e.wrapLongLines(content, e.template.Markdown.MaxLineLength)
	}

	// Apply folding for long files if enabled
	if e.template.Markdown.FoldLongFiles && len(strings.Split(content, "\n")) > 100 {
		lines := strings.Split(content, "\n")
		if len(lines) > 100 {
			content = strings.Join(lines[:50], "\n") + "\n\n<!-- ... " +
				fmt.Sprintf("%d lines truncated", len(lines)-100) + " ... -->\n\n" +
				strings.Join(lines[len(lines)-50:], "\n")
		}
	}

	return content, nil
}

// RenderMetadata renders metadata based on template options.
func (e *Engine) RenderMetadata() (string, error) {
	if !e.hasAnyMetadata() {
		return "", nil
	}

	var buf bytes.Buffer

	if e.template.Format == "markdown" {
		buf.WriteString("## Metadata\n\n")
	}

	if e.template.Metadata.IncludeTimestamp {
		buf.WriteString(fmt.Sprintf("**Generated**: %s\n", e.context.Timestamp.Format("2006-01-02 15:04:05")))
	}

	if e.template.Metadata.IncludeSourcePath {
		buf.WriteString(fmt.Sprintf("**Source**: %s\n", e.context.SourcePath))
	}

	if e.template.Metadata.IncludeFileCount {
		buf.WriteString(
			fmt.Sprintf(
				"**Files**: %d total (%d processed, %d skipped, %d errors)\n",
				e.context.TotalFiles, e.context.ProcessedFiles, e.context.SkippedFiles, e.context.ErrorFiles,
			),
		)
	}

	if e.template.Metadata.IncludeTotalSize {
		buf.WriteString(fmt.Sprintf("**Total Size**: %d bytes\n", e.context.TotalSize))
	}

	if e.template.Metadata.IncludeProcessingTime {
		buf.WriteString(fmt.Sprintf("**Processing Time**: %s\n", e.context.ProcessingTime))
	}

	if e.template.Metadata.IncludeMetrics && e.context.FilesPerSecond > 0 {
		buf.WriteString(
			fmt.Sprintf(
				"**Performance**: %.1f files/sec, %.1f MB/sec\n",
				e.context.FilesPerSecond, e.context.BytesPerSecond/1024/1024,
			),
		)
	}

	if e.template.Metadata.IncludeFileTypes && len(e.context.FileTypes) > 0 {
		buf.WriteString("**File Types**:\n")
		for fileType, count := range e.context.FileTypes {
			buf.WriteString(fmt.Sprintf("- %s: %d files\n", fileType, count))
		}
	}

	buf.WriteString("\n")

	return buf.String(), nil
}

// RenderTableOfContents renders a table of contents if enabled.
func (e *Engine) RenderTableOfContents(files []FileContext) (string, error) {
	if !e.template.Markdown.TableOfContents {
		return "", nil
	}

	var buf bytes.Buffer
	buf.WriteString("## Table of Contents\n\n")

	for _, file := range files {
		// Create markdown anchor from file path
		anchor := strings.ToLower(strings.ReplaceAll(file.RelativePath, "/", "-"))
		anchor = strings.ReplaceAll(anchor, ".", "")
		anchor = strings.ReplaceAll(anchor, " ", "-")

		buf.WriteString(fmt.Sprintf("- [%s](#%s)\n", file.RelativePath, anchor))
	}

	buf.WriteString("\n")

	return buf.String(), nil
}

// GetTemplate returns the current template.
func (e *Engine) GetTemplate() OutputTemplate {
	return e.template
}

// GetContext returns the current context.
func (e *Engine) GetContext() TemplateContext {
	return e.context
}

// renderTemplate renders a template string with the given context.
func (e *Engine) renderTemplate(templateStr string, data any) (string, error) {
	if templateStr == "" {
		return "", nil
	}

	tmpl, err := template.New("template").Funcs(e.getTemplateFunctions()).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// getTemplateFunctions returns custom functions available in templates.
func (e *Engine) getTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"formatSize": func(size int64) string {
			return e.formatBytes(size)
		},
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"basename": filepath.Base,
		"ext":      filepath.Ext,
		"dir":      filepath.Dir,
		"upper":    strings.ToUpper,
		"lower":    strings.ToLower,
		"title": func(s string) string {
			return cases.Title(language.English).String(s)
		},
		"trim": strings.TrimSpace,
		"replace": func(old, replacement, str string) string {
			return strings.ReplaceAll(str, old, replacement)
		},
		"join": strings.Join,
		"split": func(sep, str string) []string {
			return strings.Split(str, sep)
		},
	}
}

// formatBytes formats byte counts in human-readable format.
func (e *Engine) formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}

	exp := 0
	for n := bytes / unit; n >= unit; n /= unit {
		exp++
	}

	divisor := int64(1)
	for i := 0; i < exp+1; i++ {
		divisor *= 1024
	}

	return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(divisor), "KMGTPE"[exp])
}

// wrapLongLines wraps lines that exceed the specified length.
func (e *Engine) wrapLongLines(content string, maxLength int) string {
	lines := strings.Split(content, "\n")
	var wrappedLines []string

	for _, line := range lines {
		wrappedLines = append(wrappedLines, e.wrapSingleLine(line, maxLength)...)
	}

	return strings.Join(wrappedLines, "\n")
}

// wrapSingleLine wraps a single line if it exceeds maxLength.
func (e *Engine) wrapSingleLine(line string, maxLength int) []string {
	if len(line) <= maxLength {
		return []string{line}
	}

	return e.wrapLongLineWithWords(line, maxLength)
}

// wrapLongLineWithWords wraps a long line by breaking it into words.
func (e *Engine) wrapLongLineWithWords(line string, maxLength int) []string {
	words := strings.Fields(line)
	var wrappedLines []string
	var currentLine strings.Builder

	for _, word := range words {
		if e.wouldExceedLength(&currentLine, word, maxLength) {
			if currentLine.Len() > 0 {
				wrappedLines = append(wrappedLines, currentLine.String())
				currentLine.Reset()
			}
		}

		e.addWordToLine(&currentLine, word)
	}

	if currentLine.Len() > 0 {
		wrappedLines = append(wrappedLines, currentLine.String())
	}

	return wrappedLines
}

// wouldExceedLength checks if adding a word would exceed the maximum length.
func (e *Engine) wouldExceedLength(currentLine *strings.Builder, word string, maxLength int) bool {
	return currentLine.Len()+len(word)+1 > maxLength
}

// addWordToLine adds a word to the current line with appropriate spacing.
func (e *Engine) addWordToLine(currentLine *strings.Builder, word string) {
	if currentLine.Len() > 0 {
		currentLine.WriteString(" ")
	}
	currentLine.WriteString(word)
}

// hasAnyMetadata checks if any metadata options are enabled.
func (e *Engine) hasAnyMetadata() bool {
	m := e.template.Metadata

	return m.IncludeStats || m.IncludeTimestamp || m.IncludeFileCount ||
		m.IncludeSourcePath || m.IncludeFileTypes || m.IncludeProcessingTime ||
		m.IncludeTotalSize || m.IncludeMetrics
}

// ListBuiltinTemplates returns a list of available builtin templates.
func ListBuiltinTemplates() []string {
	names := make([]string, 0, len(BuiltinTemplates))
	for name := range BuiltinTemplates {
		names = append(names, name)
	}

	return names
}

// GetBuiltinTemplate returns a builtin template by name.
func GetBuiltinTemplate(name string) (OutputTemplate, bool) {
	tmpl, exists := BuiltinTemplates[name]

	return tmpl, exists
}
