# Basic Usage Examples

This directory contains practical examples of how to use gibidify for various use cases.

## Simple Code Aggregation

The most basic use case - aggregate all code files from a project into a single output:

```bash
# Aggregate all files from current directory to markdown
gibidify -source . -format markdown -destination output.md

# Aggregate specific directory to JSON
gibidify -source ./src -format json -destination code-dump.json

# Aggregate with custom worker count
gibidify -source ./project -format yaml -destination project.yaml -concurrency 8
```

## With Configuration File

For repeatable processing with custom settings:

1. Copy the configuration example:
```bash
cp config.example.yaml ~/.config/gibidify/config.yaml
```

2. Edit the configuration file to your needs, then run:
```bash
gibidify -source ./my-project
```

## Output Formats

### JSON Output
Best for programmatic processing and data analysis:

```bash
gibidify -source ./src -format json -destination api-code.json
```

Example JSON structure:
```json
{
  "files": [
    {
      "path": "src/main.go", 
      "content": "package main...",
      "language": "go",
      "size": 1024
    }
  ],
  "metadata": {
    "total_files": 15,
    "total_size": 45678,
    "processing_time": "1.2s"
  }
}
```

### Markdown Output  
Great for documentation and code reviews:

```bash
gibidify -source ./src -format markdown -destination code-review.md
```

### YAML Output
Structured and human-readable:

```bash
gibidify -source ./config -format yaml -destination config-dump.yaml
```

## Advanced Usage Examples

### Large Codebase Processing
For processing large projects with performance optimizations:

```bash
gibidify -source ./large-project \
        -format json \
        -destination large-output.json \
        -concurrency 16 \
        --verbose
```

### Memory-Conscious Processing
For systems with limited memory:

```bash
gibidify -source ./project \
        -format markdown \
        -destination output.md \
        -concurrency 4
```

### Filtered Processing
Process only specific file types (when configured):

```bash
# Configure file patterns in config.yaml
filePatterns:
  - "*.go"
  - "*.py"
  - "*.js"

# Then run
gibidify -source ./mixed-project -destination filtered.json
```

### CI/CD Integration
For automated documentation generation:

```bash
# In your CI pipeline
gibidify -source . \
        -format markdown \
        -destination docs/codebase.md \
        --no-colors \
        --no-progress \
        -concurrency 2
```

## Error Handling

### Graceful Failure Handling
The tool handles common issues gracefully:

```bash
# This will fail gracefully if source doesn't exist
gibidify -source ./nonexistent -destination out.json

# This will warn about permission issues but continue
gibidify -source ./restricted-dir -destination out.md --verbose
```

### Resource Limits
Configure resource limits in your config file:

```yaml
resourceLimits:
  enabled: true
  maxFiles: 5000
  maxTotalSize: 1073741824  # 1GB
  fileProcessingTimeoutSec: 30
  overallTimeoutSec: 1800   # 30 minutes
  hardMemoryLimitMB: 512
```

## Performance Tips

1. **Adjust Concurrency**: Start with number of CPU cores, adjust based on I/O vs CPU bound work
2. **Use Appropriate Format**: JSON is fastest, Markdown has more overhead
3. **Configure File Limits**: Set reasonable limits in config.yaml for your use case
4. **Monitor Memory**: Use `--verbose` to see memory usage during processing
5. **Use Progress Indicators**: Enable progress bars for long-running operations

## Integration Examples

### With Git Hooks
Create a pre-commit hook to generate code documentation:

```bash
#!/bin/sh
# .git/hooks/pre-commit
gibidify -source . -format markdown -destination docs/current-code.md
git add docs/current-code.md
```

### With Make
Add to your Makefile:

```makefile
.PHONY: code-dump
code-dump:
  gibidify -source ./src -format json -destination dist/codebase.json

.PHONY: docs
docs:  
  gibidify -source . -format markdown -destination docs/codebase.md
```

### Docker Usage
```dockerfile
FROM golang:1.25-alpine
RUN go install github.com/ivuorinen/gibidify@latest
WORKDIR /workspace
COPY . .
RUN gibidify -source . -format json -destination /output/codebase.json
```

## Common Use Cases

### 1. Code Review Preparation
```bash
gibidify -source ./feature-branch -format markdown -destination review.md
```

### 2. AI Code Analysis
```bash
gibidify -source ./src -format json -destination ai-input.json
```

### 3. Documentation Generation
```bash
gibidify -source ./lib -format markdown -destination api-docs.md
```

### 4. Backup Creation  
```bash
gibidify -source ./project -format yaml -destination backup-$(date +%Y%m%d).yaml
```

### 5. Code Migration Prep
```bash
gibidify -source ./legacy-code -format json -destination migration-analysis.json
```
