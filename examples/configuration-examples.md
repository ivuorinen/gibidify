# Configuration Examples

This document provides practical configuration examples for different use cases.

## Basic Configuration

Create `~/.config/gibidify/config.yaml`:

```yaml
# Basic setup for most projects
fileSizeLimit: 5242880  # 5MB per file
maxConcurrency: 8

ignoreDirectories:
  - vendor
  - node_modules
  - .git
  - dist
  - target

# Enable file type detection  
fileTypes:
  enabled: true
```

## Development Environment Configuration

Optimized for active development with fast feedback:

```yaml
# ~/.config/gibidify/config.yaml
fileSizeLimit: 1048576  # 1MB - smaller files for faster processing

ignoreDirectories:
  - vendor
  - node_modules
  - .git
  - dist
  - build
  - tmp
  - cache
  - .vscode
  - .idea

# Conservative resource limits for development
resourceLimits:
  enabled: true
  maxFiles: 1000
  maxTotalSize: 104857600  # 100MB
  fileProcessingTimeoutSec: 10
  overallTimeoutSec: 300   # 5 minutes
  maxConcurrentReads: 4
  hardMemoryLimitMB: 256

# Fast backpressure for responsive development
backpressure:
  enabled: true
  maxPendingFiles: 500
  maxPendingWrites: 50
  maxMemoryUsage: 52428800  # 50MB
  memoryCheckInterval: 100

# Simple output for quick reviews
output:
  metadata:
    includeStats: true
    includeTimestamp: true
```

## Production/CI Configuration

High-performance setup for automated processing:

```yaml
# Production configuration
fileSizeLimit: 10485760  # 10MB per file
maxConcurrency: 16

ignoreDirectories:
  - vendor
  - node_modules
  - .git
  - dist
  - build
  - target
  - tmp
  - cache
  - coverage
  - .nyc_output
  - __pycache__

# High-performance resource limits
resourceLimits:
  enabled: true
  maxFiles: 50000
  maxTotalSize: 10737418240  # 10GB
  fileProcessingTimeoutSec: 60
  overallTimeoutSec: 7200    # 2 hours
  maxConcurrentReads: 20
  hardMemoryLimitMB: 2048

# High-throughput backpressure
backpressure:
  enabled: true
  maxPendingFiles: 5000
  maxPendingWrites: 500
  maxMemoryUsage: 1073741824  # 1GB
  memoryCheckInterval: 1000

# Comprehensive output for analysis
output:
  metadata:
    includeStats: true
    includeTimestamp: true
    includeFileCount: true
    includeSourcePath: true
    includeFileTypes: true
    includeProcessingTime: true
    includeTotalSize: true
    includeMetrics: true
```

## Security-Focused Configuration

Restrictive settings for untrusted input:

```yaml
# Security-first configuration
fileSizeLimit: 1048576  # 1MB maximum

ignoreDirectories:
  - "**/.*"  # All hidden directories
  - vendor
  - node_modules
  - tmp
  - temp
  - cache

# Strict resource limits
resourceLimits:
  enabled: true
  maxFiles: 100              # Very restrictive
  maxTotalSize: 10485760     # 10MB total
  fileProcessingTimeoutSec: 5
  overallTimeoutSec: 60      # 1 minute max
  maxConcurrentReads: 2
  rateLimitFilesPerSec: 10   # Rate limiting enabled
  hardMemoryLimitMB: 128     # Low memory limit

# Conservative backpressure
backpressure:
  enabled: true
  maxPendingFiles: 50
  maxPendingWrites: 10
  maxMemoryUsage: 10485760   # 10MB
  memoryCheckInterval: 10    # Frequent checks

# Minimal file type detection
fileTypes:
  enabled: true
  # Disable potentially risky file types
  disabledLanguageExtensions:
    - .bat
    - .cmd
    - .ps1
    - .sh
  disabledBinaryExtensions:
    - .exe
    - .dll
    - .so
```

## Language-Specific Configuration

### Go Projects
```yaml
fileSizeLimit: 5242880

ignoreDirectories:
  - vendor
  - .git
  - bin
  - pkg

fileTypes:
  enabled: true
  customLanguages:
    .mod: go-mod
    .sum: go-sum

filePatterns:
  - "*.go"
  - "go.mod"
  - "go.sum"
  - "*.md"
```

### JavaScript/Node.js Projects
```yaml
fileSizeLimit: 2097152  # 2MB

ignoreDirectories:
  - node_modules
  - .git
  - dist
  - build
  - coverage
  - .nyc_output

fileTypes:
  enabled: true
  customLanguages:
    .vue: vue
    .svelte: svelte
    .astro: astro

filePatterns:
  - "*.js"
  - "*.ts"
  - "*.jsx"
  - "*.tsx"
  - "*.vue"
  - "*.json"
  - "*.md"
```

### Python Projects  
```yaml
fileSizeLimit: 5242880

ignoreDirectories:
  - .git
  - __pycache__
  - .pytest_cache
  - venv
  - env
  - .env
  - dist
  - build
  - .tox

fileTypes:
  enabled: true
  customLanguages:
    .pyi: python-interface
    .ipynb: jupyter-notebook

filePatterns:
  - "*.py"
  - "*.pyi" 
  - "requirements*.txt"
  - "*.toml"
  - "*.cfg"
  - "*.ini"
  - "*.md"
```

## Output Format Configurations

### Detailed Markdown Output
```yaml
output:
  template: "detailed"
  
  metadata:
    includeStats: true
    includeTimestamp: true
    includeFileCount: true
    includeSourcePath: true
    includeFileTypes: true
    includeProcessingTime: true
    
  markdown:
    useCodeBlocks: true
    includeLanguage: true
    headerLevel: 2
    tableOfContents: true
    syntaxHighlighting: true
    lineNumbers: true
    maxLineLength: 120
    
  variables:
    project_name: "My Project"
    author: "Development Team"
    version: "1.0.0"
```

### Compact JSON Output
```yaml
output:
  template: "minimal"
  
  metadata:
    includeStats: true
    includeFileCount: true
```

### Custom Template Output
```yaml  
output:
  template: "custom"
  
  custom:
    header: |
      # {{ .ProjectName }} Code Dump
      Generated: {{ .Timestamp }}
      Total Files: {{ .FileCount }}
      
    footer: |
      ---
      Processing completed in {{ .ProcessingTime }}
      
    fileHeader: |
      ## {{ .Path }}
      Language: {{ .Language }} | Size: {{ .Size }} bytes
      
    fileFooter: ""
    
  variables:
    project_name: "Custom Project"
```

## Environment-Specific Configurations

### Docker Container
```yaml
# Optimized for containerized environments
fileSizeLimit: 5242880
maxConcurrency: 4  # Conservative for containers

resourceLimits:
  enabled: true
  hardMemoryLimitMB: 512
  maxFiles: 5000
  overallTimeoutSec: 1800

backpressure:
  enabled: true
  maxMemoryUsage: 268435456  # 256MB
```

### GitHub Actions
```yaml
# CI/CD optimized configuration  
fileSizeLimit: 2097152
maxConcurrency: 2  # Conservative for shared runners

ignoreDirectories:
  - .git
  - .github
  - node_modules
  - vendor
  - dist
  - build

resourceLimits:
  enabled: true
  maxFiles: 2000
  overallTimeoutSec: 900  # 15 minutes
  hardMemoryLimitMB: 1024
```

### Local Development
```yaml
# Developer-friendly settings
fileSizeLimit: 10485760  # 10MB
maxConcurrency: 8

# Show progress and verbose output
output:
  metadata:
    includeStats: true
    includeTimestamp: true
    includeProcessingTime: true
    includeMetrics: true
    
  markdown:
    useCodeBlocks: true
    includeLanguage: true
    syntaxHighlighting: true
```

## Template Examples

### Custom API Documentation Template
```yaml
output:
  template: "custom"
  
  custom:
    header: |
      # {{ .Variables.api_name }} API Documentation
      Version: {{ .Variables.version }}
      Generated: {{ .Timestamp }}
      
      ## Overview
      This document contains the complete source code for the {{ .Variables.api_name }} API.
      
      ## Statistics
      - Total Files: {{ .FileCount }}
      - Total Size: {{ .TotalSize | formatSize }}
      - Processing Time: {{ .ProcessingTime }}
      
      ---
      
    fileHeader: |
      ### {{ .Path }}
      
      **Type:** {{ .Language }}  
      **Size:** {{ .Size | formatSize }}
      
      ```{{ .Language }}
      
    fileFooter: |
      ```
      
      ---
      
    footer: |
      ## Summary
      
      Documentation generated with [gibidify](https://github.com/ivuorinen/gibidify)
      
  variables:
    api_name: "My API"
    version: "v1.2.3"
```

### Code Review Template
```yaml
output:
  template: "custom"
  
  custom:
    header: |
      # Code Review: {{ .Variables.pr_title }}
      
      **PR Number:** #{{ .Variables.pr_number }}  
      **Author:** {{ .Variables.author }}  
      **Date:** {{ .Timestamp }}
      
      ## Files Changed ({{ .FileCount }})
      
    fileHeader: |
      ## 📄 {{ .Path }}
      
      <details>
      <summary>{{ .Language | upper }} • {{ .Size | formatSize }}</summary>
      
      ```{{ .Language }}
      
    fileFooter: |
      ```
      
      </details>
      
    footer: |
      ---
      
      **Review Summary:**
      - Files reviewed: {{ .FileCount }}
      - Total size: {{ .TotalSize | formatSize }}
      - Generated in: {{ .ProcessingTime }}
      
  variables:
    pr_title: "Feature Implementation"
    pr_number: "123"
    author: "developer@example.com"
```
