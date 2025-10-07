package cli

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ivuorinen/gibidify/fileproc"
)

func TestProcessorSimple(t *testing.T) {
	t.Run("NewProcessor", func(t *testing.T) {
		flags := &Flags{
			SourceDir:   "/tmp/test",
			Destination: "output.md",
			Format:      "markdown",
			Concurrency: 2,
			NoColors:    true,
			NoProgress:  true,
			Verbose:     false,
		}

		p := NewProcessor(flags)

		assert.NotNil(t, p)
		assert.Equal(t, flags, p.flags)
		assert.NotNil(t, p.ui)
		assert.NotNil(t, p.backpressure)
		assert.NotNil(t, p.resourceMonitor)
		assert.False(t, p.ui.enableColors)
		assert.False(t, p.ui.enableProgress)
	})

	t.Run("ConfigureFileTypes", func(t *testing.T) {
		p := &Processor{
			flags: &Flags{},
			ui:    NewUIManager(),
		}

		// Should not panic or error
		err := p.configureFileTypes()
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("CreateOutputFile", func(t *testing.T) {
		// Create temp file path
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "output.txt")

		p := &Processor{
			flags: &Flags{
				Destination: outputPath,
			},
			ui: NewUIManager(),
		}

		file, err := p.createOutputFile()
		assert.NoError(t, err)
		assert.NotNil(t, file)

		// Clean up
		err = file.Close()
		require.NoError(t, err)
		err = os.Remove(outputPath)
		require.NoError(t, err)
	})

	t.Run("ValidateFileCollection", func(t *testing.T) {
		p := &Processor{
			ui: NewUIManager(),
		}

		// Empty collection should be valid (just checks limits)
		err := p.validateFileCollection([]string{})
		assert.NoError(t, err)

		// Small collection should be valid
		err = p.validateFileCollection([]string{
			"/test/file1.go",
			"/test/file2.go",
		})
		assert.NoError(t, err)
	})

	t.Run("CollectFiles_EmptyDir", func(t *testing.T) {
		tempDir := t.TempDir()

		p := &Processor{
			flags: &Flags{
				SourceDir: tempDir,
			},
			ui: NewUIManager(),
		}

		files, err := p.collectFiles()
		assert.NoError(t, err)
		assert.Empty(t, files)
	})

	t.Run("CollectFiles_WithFiles", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create test files
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "test1.go"), []byte("package main"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "test2.go"), []byte("package test"), 0o644))

		// Set config so no files are ignored, and restore after test
		origIgnoreDirs := viper.Get("ignoreDirectories")
		origFileSizeLimit := viper.Get("fileSizeLimit")
		viper.Set("ignoreDirectories", []string{})
		viper.Set("fileSizeLimit", 1024*1024*10) // 10MB
		t.Cleanup(func() {
			viper.Set("ignoreDirectories", origIgnoreDirs)
			viper.Set("fileSizeLimit", origFileSizeLimit)
		})

		p := &Processor{
			flags: &Flags{
				SourceDir: tempDir,
			},
			ui: NewUIManager(),
		}

		files, err := p.collectFiles()
		assert.NoError(t, err)
		assert.Len(t, files, 2)
	})

	t.Run("SendFiles", func(t *testing.T) {
		p := &Processor{
			backpressure: fileproc.NewBackpressureManager(),
			ui:           NewUIManager(),
		}

		ctx := context.Background()
		fileCh := make(chan string, 3)
		files := []string{
			"/test/file1.go",
			"/test/file2.go",
		}

		var wg sync.WaitGroup
		wg.Add(1)
		// Send files in a goroutine since it might block
		go func() {
			defer wg.Done()
			err := p.sendFiles(ctx, files, fileCh)
			assert.NoError(t, err)
		}()

		// Read all files from channel
		var received []string
		for i := 0; i < len(files); i++ {
			file := <-fileCh
			received = append(received, file)
		}

		assert.Equal(t, len(files), len(received))

		// Wait for sendFiles goroutine to finish (and close fileCh)
		wg.Wait()

		// Now channel should be closed
		_, ok := <-fileCh
		assert.False(t, ok, "channel should be closed")
	})

	t.Run("WaitForCompletion", func(t *testing.T) {
		p := &Processor{
			ui: NewUIManager(),
		}

		writeCh := make(chan fileproc.WriteRequest)
		writerDone := make(chan struct{})

		// Simulate writer finishing
		go func() {
			<-writeCh // Wait for close
			close(writerDone)
		}()

		var wg sync.WaitGroup
		// Start and finish immediately
		wg.Add(1)
		wg.Done()

		// Should complete without hanging
		p.waitForCompletion(&wg, writeCh, writerDone)
		assert.NotNil(t, p)
	})

	t.Run("LogFinalStats", func(t *testing.T) {
		p := &Processor{
			flags: &Flags{
				Verbose: true,
			},
			ui:              NewUIManager(),
			resourceMonitor: fileproc.NewResourceMonitor(),
			backpressure:    fileproc.NewBackpressureManager(),
		}

		// Should not panic
		p.logFinalStats()
		assert.NotNil(t, p)
	})
}

// Test error handling scenarios
func TestProcessorErrors(t *testing.T) {
	t.Run("CreateOutputFile_InvalidPath", func(t *testing.T) {
		p := &Processor{
			flags: &Flags{
				Destination: "/root/cannot-write-here.txt",
			},
			ui: NewUIManager(),
		}

		file, err := p.createOutputFile()
		assert.Error(t, err)
		assert.Nil(t, file)
	})

	t.Run("CollectFiles_NonExistentDir", func(t *testing.T) {
		p := &Processor{
			flags: &Flags{
				SourceDir: "/non/existent/path",
			},
			ui: NewUIManager(),
		}

		files, err := p.collectFiles()
		assert.Error(t, err)
		assert.Nil(t, files)
	})

	t.Run("SendFiles_WithCancellation", func(t *testing.T) {
		p := &Processor{
			backpressure: fileproc.NewBackpressureManager(),
			ui:           NewUIManager(),
		}

		ctx, cancel := context.WithCancel(context.Background())
		fileCh := make(chan string) // Unbuffered to force blocking

		files := []string{
			"/test/file1.go",
			"/test/file2.go",
			"/test/file3.go",
		}

		// Cancel immediately
		cancel()

		err := p.sendFiles(ctx, files, fileCh)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})
}
