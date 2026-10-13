# Configuration Examples

Practical `config.yaml` examples. gibidify reads the first file found at:

- `$XDG_CONFIG_HOME/gibidify/config.yaml` (only when `XDG_CONFIG_HOME` is an
  absolute path)
- `$HOME/.config/gibidify/config.yaml`

Every available key appears below; `config.example.yaml` documents defaults and
valid ranges. Unknown keys are reported as a warning and ignored. A file that
cannot be parsed, or any value that fails validation, stops gibidify with exit
status 1 instead of running on the defaults.

`ignoreDirectories` entries are plain directory names (no `/`), matched at any
depth; dot-directories such as `.venv` are allowed. `.gitignore` and `.ignore`
files are honoured as well, and `.git`, `.hg` and `.svn` are always skipped, so
they need no entry.

## Basic Configuration

```yaml
fileSizeLimit: 5242880 # 5MB per file

ignoreDirectories:
  - vendor
  - node_modules
  - dist
  - target
```

## Small, Fast Runs

```yaml
fileSizeLimit: 1048576 # 1MB per file
maxConcurrency: 4 # ceiling for -concurrency

resourceLimits:
  maxFiles: 1000
  maxTotalSize: 104857600 # 100MB
```

## Large Repositories

```yaml
fileSizeLimit: 10485760 # 10MB per file

ignoreDirectories:
  - vendor
  - node_modules
  - dist
  - build
  - target
  - coverage
  - __pycache__

resourceLimits:
  maxFiles: 50000
  maxTotalSize: 10737418240 # 10GB
```

## Shared CI Runners

Cap the worker count so a large `-concurrency` (or the default, one worker per
CPU) cannot exceed what the runner should use:

```yaml
maxConcurrency: 2

resourceLimits:
  maxFiles: 2000
```

## File Type Customization

```yaml
fileTypes:
  enabled: true
  customImageExtensions:
    - .avif
  customBinaryExtensions:
    - .dat
  customLanguages:
    .zig: zig
    .vue: vue
  disabledImageExtensions:
    - .svg # include SVG source in the output
  disabledBinaryExtensions:
    - .bin
  disabledLanguageExtensions:
    - .bat
```

Files with a NUL byte in their first 8000 bytes are always skipped as binary,
whatever their extension.

## Language-Specific Configuration

### Go Projects

```yaml
ignoreDirectories:
  - vendor
  - bin

fileTypes:
  enabled: true
  customLanguages:
    .mod: go-mod
    .sum: go-sum
```

### JavaScript/Node.js Projects

```yaml
fileSizeLimit: 2097152 # 2MB

ignoreDirectories:
  - node_modules
  - dist
  - build
  - coverage

fileTypes:
  enabled: true
  customLanguages:
    .vue: vue
    .svelte: svelte
    .astro: astro
```

### Python Projects

```yaml
ignoreDirectories:
  - __pycache__
  - venv
  - .venv
  - .tox
  - dist
  - build

fileTypes:
  enabled: true
  customLanguages:
    .pyi: python
```
