// Package cli provides command-line interface functionality for gibidify.
package cli

import (
	"fmt"
	"os"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/shared"
)

// collectFiles collects all files to be processed.
func (p *Processor) collectFiles() ([]string, error) {
	files, err := fileproc.CollectFiles(p.flags.SourceDir)
	if err != nil {
		return nil, shared.WrapError(
			err,
			shared.ErrorTypeProcessing,
			shared.CodeProcessingCollection,
			"error collecting files",
		)
	}

	logger := shared.GetLogger()
	logger.Infof(shared.CLIMsgFoundFilesToProcess, len(files))

	return files, nil
}

// validateFileCollection validates the collected files against resource limits.
func (p *Processor) validateFileCollection(files []string) error {
	if !config.ResourceLimitsEnabled() {
		return nil
	}

	// Check file count limit
	maxFiles := config.MaxFiles()
	if len(files) > maxFiles {
		return shared.NewStructuredError(
			shared.ErrorTypeValidation,
			shared.CodeResourceLimitFiles,
			fmt.Sprintf("file count (%d) exceeds maximum limit (%d)", len(files), maxFiles),
			"",
			map[string]any{
				"file_count": len(files),
				"max_files":  maxFiles,
			},
		)
	}

	// Check total size limit (estimate)
	maxTotalSize := config.MaxTotalSize()
	totalSize := int64(0)
	oversizedFiles := 0

	for _, filePath := range files {
		if fileInfo, err := os.Stat(filePath); err == nil {
			totalSize += fileInfo.Size()
			if totalSize > maxTotalSize {
				return shared.NewStructuredError(
					shared.ErrorTypeValidation,
					shared.CodeResourceLimitTotalSize,
					fmt.Sprintf(
						"total file size (%d bytes) would exceed maximum limit (%d bytes)", totalSize, maxTotalSize,
					),
					"",
					map[string]any{
						"total_size":     totalSize,
						"max_total_size": maxTotalSize,
						"files_checked":  len(files),
					},
				)
			}
		} else {
			oversizedFiles++
		}
	}

	logger := shared.GetLogger()
	if oversizedFiles > 0 {
		logger.Warnf("Could not stat %d files during pre-validation", oversizedFiles)
	}

	logger.Infof("Pre-validation passed: %d files, %d MB total", len(files), totalSize/int64(shared.BytesPerMB))

	return nil
}
