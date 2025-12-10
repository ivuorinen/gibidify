package templates

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ivuorinen/gibidify/shared"
)

func TestNewEngine(t *testing.T) {
	context := TemplateContext{
		Timestamp:  time.Now(),
		SourcePath: "/test/source",
		Format:     "markdown",
	}

	engine, err := NewEngine("default", context)
	if err != nil {
		t.Fatalf(shared.TestMsgNewEngineFailed, err)
	}

	if engine == nil {
		t.Fatal("Engine is nil")
	}

	if engine.template.Name != "Default" {
		t.Errorf("Expected template name 'Default', got '%s'", engine.template.Name)
	}
}

func TestNewEngineUnknownTemplate(t *testing.T) {
	context := TemplateContext{}

	_, err := NewEngine("nonexistent", context)
	if err == nil {
		t.Error("Expected error for unknown template")
	}

	if !strings.Contains(err.Error(), "template 'nonexistent' not found") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestNewEngineWithCustomTemplate(t *testing.T) {
	customTemplate := OutputTemplate{
		Name:        "Custom",
		Description: "Custom template",
		Format:      "markdown",
		Header:      "# Custom Header",
		Footer:      "Custom Footer",
	}

	context := TemplateContext{
		SourcePath: "/test",
	}

	engine := NewEngineWithCustomTemplate(customTemplate, context)

	if engine == nil {
		t.Fatal("Engine is nil")
	}

	if engine.template.Name != "Custom" {
		t.Errorf("Expected template name 'Custom', got '%s'", engine.template.Name)
	}
}

func TestRenderHeader(t *testing.T) {
	context := TemplateContext{
		Timestamp:  time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC),
		SourcePath: shared.TestPathTestProject,
		Format:     "markdown",
	}

	engine, err := NewEngine("default", context)
	if err != nil {
		t.Fatalf(shared.TestMsgNewEngineFailed, err)
	}

	header, err := engine.RenderHeader()
	if err != nil {
		t.Fatalf("RenderHeader failed: %v", err)
	}

	if !strings.Contains(header, shared.TestPathTestProject) {
		t.Errorf("Header should contain source path, got: %s", header)
	}

	if !strings.Contains(header, "2023-12-25") {
		t.Errorf("Header should contain timestamp, got: %s", header)
	}
}

func TestRenderFooter(t *testing.T) {
	engine, err := NewEngine("default", TemplateContext{})
	if err != nil {
		t.Fatalf(shared.TestMsgNewEngineFailed, err)
	}

	footer, err := engine.RenderFooter()
	if err != nil {
		t.Fatalf("RenderFooter failed: %v", err)
	}

	if !strings.Contains(footer, "gibidify") {
		t.Errorf("Footer should contain 'gibidify', got: %s", footer)
	}
}

func TestRenderFileHeader(t *testing.T) {
	engine, err := NewEngine("default", TemplateContext{})
	if err != nil {
		t.Fatalf(shared.TestMsgNewEngineFailed, err)
	}

	fileCtx := FileContext{
		Path:         shared.TestPathTestFileGo,
		RelativePath: shared.TestFileGoExt,
		Name:         shared.TestFileGoExt,
		Language:     "go",
		Size:         1024,
	}

	header, err := engine.RenderFileHeader(fileCtx)
	if err != nil {
		t.Fatalf("RenderFileHeader failed: %v", err)
	}

	if !strings.Contains(header, shared.TestFileGoExt) {
		t.Errorf("File header should contain filename, got: %s", header)
	}

	if !strings.Contains(header, "```go") {
		t.Errorf("File header should contain language code block, got: %s", header)
	}
}

func TestRenderFileFooter(t *testing.T) {
	engine, err := NewEngine("default", TemplateContext{})
	if err != nil {
		t.Fatalf(shared.TestMsgNewEngineFailed, err)
	}

	fileCtx := FileContext{
		Path: shared.TestPathTestFileGo,
	}

	footer, err := engine.RenderFileFooter(fileCtx)
	if err != nil {
		t.Fatalf("RenderFileFooter failed: %v", err)
	}

	if !strings.Contains(footer, "```") {
		t.Errorf("File footer should contain code block close, got: %s", footer)
	}
}

func TestRenderFileContentBasic(t *testing.T) {
	engine, err := NewEngine("default", TemplateContext{})
	if err != nil {
		t.Fatalf(shared.TestMsgNewEngineFailed, err)
	}

	fileCtx := FileContext{
		Content: "package main\n\nfunc main() {\n    fmt.Println(\"hello\")\n}",
	}

	content, err := engine.RenderFileContent(fileCtx)
	if err != nil {
		t.Fatalf(shared.TestMsgRenderFileContentFailed, err)
	}

	if content != fileCtx.Content {
		t.Errorf("Content should be unchanged for basic case, got: %s", content)
	}
}

func TestRenderFileContentLongLines(t *testing.T) {
	customTemplate := OutputTemplate{
		Format: "markdown",
		Markdown: MarkdownOptions{
			MaxLineLength: 20,
		},
	}

	engine := NewEngineWithCustomTemplate(customTemplate, TemplateContext{})

	fileCtx := FileContext{
		Content: "this is a very long line that should be wrapped",
	}

	content, err := engine.RenderFileContent(fileCtx)
	if err != nil {
		t.Fatalf(shared.TestMsgRenderFileContentFailed, err)
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if len(line) > 20 {
			t.Errorf("Line length should not exceed 20 characters, got line: %s (len=%d)", line, len(line))
		}
	}
}

func TestRenderFileContentFoldLongFiles(t *testing.T) {
	customTemplate := OutputTemplate{
		Format: "markdown",
		Markdown: MarkdownOptions{
			FoldLongFiles: true,
		},
	}

	engine := NewEngineWithCustomTemplate(customTemplate, TemplateContext{})

	// Create content with more than 100 lines
	lines := make([]string, 150)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i+1)
	}

	fileCtx := FileContext{
		Content: strings.Join(lines, "\n"),
	}

	content, err := engine.RenderFileContent(fileCtx)
	if err != nil {
		t.Fatalf(shared.TestMsgRenderFileContentFailed, err)
	}

	if !strings.Contains(content, "lines truncated") {
		t.Error("Content should contain truncation message")
	}
}

func TestRenderMetadata(t *testing.T) {
	context := TemplateContext{
		Timestamp:      time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC),
		SourcePath:     shared.TestPathTestProject,
		TotalFiles:     10,
		ProcessedFiles: 8,
		SkippedFiles:   1,
		ErrorFiles:     1,
		TotalSize:      1024000,
		ProcessingTime: "2.5s",
		FilesPerSecond: 3.2,
		BytesPerSecond: 409600,
		FileTypes: map[string]int{
			"go":   5,
			"js":   3,
			"yaml": 2,
		},
	}

	engine, err := NewEngine("detailed", context)
	if err != nil {
		t.Fatalf(shared.TestMsgNewEngineFailed, err)
	}

	metadata, err := engine.RenderMetadata()
	if err != nil {
		t.Fatalf("RenderMetadata failed: %v", err)
	}

	if !strings.Contains(metadata, "2023-12-25") {
		t.Error("Metadata should contain timestamp")
	}

	if !strings.Contains(metadata, shared.TestPathTestProject) {
		t.Error("Metadata should contain source path")
	}

	if !strings.Contains(metadata, "10 total") {
		t.Error("Metadata should contain file count")
	}

	if !strings.Contains(metadata, "1024000 bytes") {
		t.Error("Metadata should contain total size")
	}

	if !strings.Contains(metadata, "2.5s") {
		t.Error("Metadata should contain processing time")
	}

	if !strings.Contains(metadata, "3.2 files/sec") {
		t.Error("Metadata should contain performance metrics")
	}

	if !strings.Contains(metadata, "go: 5 files") {
		t.Error("Metadata should contain file types")
	}
}

func TestRenderTableOfContents(t *testing.T) {
	engine, err := NewEngine("detailed", TemplateContext{})
	if err != nil {
		t.Fatalf(shared.TestMsgNewEngineFailed, err)
	}

	files := []FileContext{
		{RelativePath: "main.go"},
		{RelativePath: "utils/helper.go"},
		{RelativePath: "config.yaml"},
	}

	toc, err := engine.RenderTableOfContents(files)
	if err != nil {
		t.Fatalf("RenderTableOfContents failed: %v", err)
	}

	if !strings.Contains(toc, "Table of Contents") {
		t.Error("TOC should contain header")
	}

	if !strings.Contains(toc, "[main.go]") {
		t.Error("TOC should contain main.go link")
	}

	if !strings.Contains(toc, "[utils/helper.go]") {
		t.Error("TOC should contain utils/helper.go link")
	}

	if !strings.Contains(toc, "[config.yaml]") {
		t.Error("TOC should contain config.yaml link")
	}
}

func TestRenderTableOfContentsDisabled(t *testing.T) {
	engine, err := NewEngine("default", TemplateContext{})
	if err != nil {
		t.Fatalf(shared.TestMsgNewEngineFailed, err)
	}

	files := []FileContext{{RelativePath: "test.go"}}

	toc, err := engine.RenderTableOfContents(files)
	if err != nil {
		t.Fatalf("RenderTableOfContents failed: %v", err)
	}

	if toc != "" {
		t.Errorf("TOC should be empty when disabled, got: %s", toc)
	}
}

func TestTemplateFunctions(t *testing.T) {
	engine, err := NewEngine("default", TemplateContext{})
	if err != nil {
		t.Fatalf(shared.TestMsgNewEngineFailed, err)
	}

	testCases := []struct {
		name     string
		template string
		context  any
		expected string
	}{
		{
			name:     "formatSize",
			template: "{{formatSize .Size}}",
			context:  struct{ Size int64 }{Size: 1024},
			expected: "1.0KB",
		},
		{
			name:     "basename",
			template: "{{basename .Path}}",
			context:  struct{ Path string }{Path: shared.TestPathTestFileGo},
			expected: shared.TestFileGoExt,
		},
		{
			name:     "ext",
			template: "{{ext .Path}}",
			context:  struct{ Path string }{Path: shared.TestPathTestFileGo},
			expected: ".go",
		},
		{
			name:     "upper",
			template: "{{upper .Text}}",
			context:  struct{ Text string }{Text: "hello"},
			expected: "HELLO",
		},
		{
			name:     "lower",
			template: "{{lower .Text}}",
			context:  struct{ Text string }{Text: "HELLO"},
			expected: "hello",
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.name, func(t *testing.T) {
				result, err := engine.renderTemplate(tc.template, tc.context)
				if err != nil {
					t.Fatalf("Template rendering failed: %v", err)
				}

				if result != tc.expected {
					t.Errorf("Expected %q, got %q", tc.expected, result)
				}
			},
		)
	}
}

func TestListBuiltinTemplates(t *testing.T) {
	templates := ListBuiltinTemplates()

	if len(templates) == 0 {
		t.Error("Should have builtin templates")
	}

	expectedTemplates := []string{"default", "minimal", "detailed", "compact"}
	for _, expected := range expectedTemplates {
		found := false
		for _, tmpl := range templates {
			if tmpl == expected {
				found = true

				break
			}
		}
		if !found {
			t.Errorf("Expected template %s not found in list", expected)
		}
	}
}

func TestBuiltinTemplate(t *testing.T) {
	tmpl, exists := BuiltinTemplate("default")
	if !exists {
		t.Error("Default template should exist")
	}

	if tmpl.Name != "Default" {
		t.Errorf("Expected name 'Default', got %s", tmpl.Name)
	}

	_, exists = BuiltinTemplate("nonexistent")
	if exists {
		t.Error("Nonexistent template should not exist")
	}
}

func TestFormatBytes(t *testing.T) {
	engine, err := NewEngine("default", TemplateContext{})
	if err != nil {
		t.Fatalf(shared.TestMsgNewEngineFailed, err)
	}

	testCases := []struct {
		bytes    int64
		expected string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1024 * 1024, "1.0MB"},
		{5 * 1024 * 1024 * 1024, "5.0GB"},
	}

	for _, tc := range testCases {
		t.Run(
			tc.expected, func(t *testing.T) {
				result := engine.formatBytes(tc.bytes)
				if result != tc.expected {
					t.Errorf("formatBytes(%d) = %s, want %s", tc.bytes, result, tc.expected)
				}
			},
		)
	}
}

// validateTemplateRendering validates all template rendering functions for a given engine.
func validateTemplateRendering(t *testing.T, engine *Engine, name string) {
	t.Helper()

	// Test header rendering
	_, err := engine.RenderHeader()
	if err != nil {
		t.Errorf("Failed to render header for template %s: %v", name, err)
	}

	// Test footer rendering
	_, err = engine.RenderFooter()
	if err != nil {
		t.Errorf("Failed to render footer for template %s: %v", name, err)
	}

	// Test file rendering
	validateFileRendering(t, engine, name)
}

// validateFileRendering validates file header and footer rendering for a given engine.
func validateFileRendering(t *testing.T, engine *Engine, name string) {
	t.Helper()

	fileCtx := FileContext{
		Path:         "/test.go",
		RelativePath: "test.go",
		Language:     "go",
		Size:         100,
	}

	// Test file header rendering
	_, err := engine.RenderFileHeader(fileCtx)
	if err != nil {
		t.Errorf("Failed to render file header for template %s: %v", name, err)
	}

	// Test file footer rendering
	_, err = engine.RenderFileFooter(fileCtx)
	if err != nil {
		t.Errorf("Failed to render file footer for template %s: %v", name, err)
	}
}

func TestBuiltinTemplatesIntegrity(t *testing.T) {
	// Test that all builtin templates are valid and can be used
	context := TemplateContext{
		Timestamp:  time.Now(),
		SourcePath: "/test",
		Format:     "markdown",
	}

	for name := range BuiltinTemplates {
		t.Run(
			name, func(t *testing.T) {
				engine, err := NewEngine(name, context)
				if err != nil {
					t.Fatalf("Failed to create engine for template %s: %v", name, err)
				}

				validateTemplateRendering(t, engine, name)
			},
		)
	}
}
