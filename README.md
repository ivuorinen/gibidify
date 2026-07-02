# gibidify

gibidify is a CLI application written in Go that scans a source directory
recursively and aggregates code files into a single text file. The output
file is designed for use with LLM models and includes a prefix,
file sections with separators, and a suffix.

## Features

- **Recursive directory scanning** with smart file filtering
- **Configurable file type detection** - add/remove extensions and languages
- **Multiple output formats** - markdown, JSON, YAML
- **Memory-optimized processing** - streaming for large files, intelligent back-pressure
- **Concurrent processing** with configurable worker pools
- **Comprehensive configuration** via YAML with validation
- **Production-ready** with structured error handling
- **Modular architecture** - clean, focused codebase
- **Enhanced CLI experience** - progress bars, colored output, helpful error messages
- **Cross-platform** with Docker support

## Installation

Clone the repository and build the application:

```bash
git clone https://github.com/ivuorinen/gibidify.git
cd gibidify
go build -o gibidify .
```

## Usage

```bash
./gibidify \
  -source <source_directory> \
  -destination <output_file> \
  -format markdown|json|yaml \
  -concurrency <num_workers> \
  --prefix="..." \
  --suffix="..." \
  --no-colors \
  --no-progress \
  --verbose \
  --log-level debug
```

Flags:

- `-source`: directory to scan.
- `-destination`: output file path (optional; defaults to `<source>.<format>`).
- `-format`: output format (`markdown`, `json`, or `yaml`).
- `-concurrency`: number of concurrent workers.
- `--prefix` / `--suffix`: optional text blocks.
- `--no-colors`: disable colored terminal output.
- `--no-progress`: disable progress bars.
- `--no-ui`: disable all UI output (implies `--no-colors` and `--no-progress`).
- `--verbose`: enable verbose output and detailed logging.
- `--log-level`: set log level (default: warn; accepted values: debug, info, warn, error).

## Docker

A Docker image can be built using the provided Dockerfile:

```bash
docker build -t ghcr.io/ivuorinen/gibidify:<tag> .
```

Run the Docker container:

```bash
docker run --rm \
  -v $(pwd):/workspace \
  -v $HOME/.config/gibidify:/config \
  ghcr.io/ivuorinen/gibidify:<tag> \
  -source /workspace/your_source_directory \
  -destination /workspace/output.txt \
  --prefix="Your prefix text" \
  --suffix="Your suffix text"
```

## Configuration

gibidify supports a YAML configuration file. Place it at:

- `$XDG_CONFIG_HOME/gibidify/config.yaml` or
- `$HOME/.config/gibidify/config.yaml` or
- in the folder you run the application from.

Example configuration:

```yaml
fileSizeLimit: 5242880  # 5 MB
ignoreDirectories:
  - vendor
  - node_modules
  - .git
  - dist
  - build
  - target

# FileType customization
fileTypes:
  enabled: true
  # Add custom file extensions
  customImageExtensions:
    - .webp
    - .avif
  customBinaryExtensions:
    - .custom
  customLanguages:
    .zig: zig
    .odin: odin
    .v: vlang
  # Disable default extensions
  disabledImageExtensions:
    - .bmp
  disabledBinaryExtensions:
    - .exe
  disabledLanguageExtensions:
    - .bat

# Safety caps
resourceLimits:
  maxFiles: 10000            # Max number of files to process
  maxTotalSize: 1073741824   # Max total bytes across all files (1GB)
```

See `config.example.yaml` for a comprehensive configuration example.

## License

This project is licensed under [the MIT License](LICENSE).
