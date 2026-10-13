# gibidify

gibidify is a CLI application written in Go that scans a source directory
recursively and aggregates code files into a single text file. The output
file is designed for use with LLM models and includes a prefix,
file sections with separators, and a suffix.

## Features

- **Recursive directory scanning** with smart file filtering
- **Configurable file type detection** - add/remove extensions and languages
- **Multiple output formats** - markdown, JSON, YAML
- **Safety caps** - configurable max file count and total size limits
- **Concurrent processing** with configurable worker pools
- **Comprehensive configuration** via YAML with validation
- **Production-ready** with structured error handling
- **Modular architecture** - clean, focused codebase
- **Enhanced CLI experience** - progress bars, colored output, helpful error
  messages
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
- `-destination`: output file path (optional; defaults to
  `<source-directory-name>.<format>` in the
  current directory, e.g. `src.markdown`). The output file is never read back as
  input, even when it
  lies inside the source tree.
- `-format`: output format (`markdown`, `json`, or `yaml`).
- `-concurrency`: number of concurrent workers.
- `--prefix` / `--suffix`: optional text blocks.
- `--no-colors`: disable colored terminal output.
- `--no-progress`: disable progress bars.
- `--no-ui`: disable all UI output (implies `--no-colors` and `--no-progress`).
- `--verbose`: shorthand for `--log-level info` (an explicit `--log-level`
  wins).
- `--log-level`: set log level (default: warn; accepted values: debug, info,
  warn, error).
- `--include-sensitive`: keep credential files that are skipped by default (see
  below).
- `--version`: print version information and exit.

Flags must come before any other argument; a stray argument is an error
rather than silently ending flag parsing.

Output order follows the directory walk, so the same input always produces the
same output. The exit status is non-zero when any file could not be read or the
output could not be written. The output is written to a temporary file and
renamed into place only when the run succeeds, so a failed or interrupted run
(Ctrl-C exits 130) leaves an existing output file untouched. A new output file
is created readable by its owner only (mode 0600); replacing an existing one
keeps that file's permissions.

### What is included

Every regular file under `-source` is included unless one of these rules
excludes it, checked in this order:

1. Symlinks and non-regular files (FIFOs, sockets, devices) are never read.
2. Version-control metadata (`.git`, `.hg`, `.svn`) is always skipped.
3. Directories named in `ignoreDirectories` (default: `vendor`,
   `node_modules`, `dist`, `build`, `target`, `bower_components`, `cache`,
   `tmp`) are skipped at any depth.
4. Paths matched by `.gitignore` and `.ignore` files at any level are skipped;
   a deeper file's `!pattern` re-includes a path a parent ignores.
5. Credential files are skipped and logged unless `--include-sensitive` is
   given: `.env`, `.env.*`, `*.pem`, `*.key`, `*.p12`, `*.pfx`, `id_rsa*`,
   `id_dsa*`, `id_ecdsa*`, `id_ed25519*`, `.npmrc`, `.netrc`, `.pgpass`.
6. Files over `fileSizeLimit` (default 5 MB) are skipped and logged.
7. Binary and image files are skipped by extension, and any file with a NUL
   byte in its first 8000 bytes is skipped as binary.

The output file itself is never collected, even when it lies inside the source
tree. Files that are not valid UTF-8 are included with the invalid bytes
replaced by U+FFFD, and a warning names them.

### Environment

- `XDG_CONFIG_HOME` / `HOME`: where the config file is looked up (see
  Configuration).
- `NO_COLOR`: disables colored output ([no-color.org](https://no-color.org/)).
- `FORCE_COLOR`: enables colored output even when stderr is not a terminal.
- `TERM=dumb` (or unset): disables colored output.
- `CI`: disables colored output, except on GitHub Actions
  (`GITHUB_ACTIONS=true`).

## Docker

A Docker image can be built from a checkout using the provided Dockerfile (it
compiles from source):

```bash
docker build -t ghcr.io/ivuorinen/gibidify:<tag> .
```

Run the Docker container. `--user` lets it write to the mounted directory, and
the config directory
is mounted where `XDG_CONFIG_HOME` points:

```bash
docker run --rm \
  --user "$(id -u):$(id -g)" \
  -v "$(pwd)":/workspace \
  -v "$HOME/.config/gibidify":/config/gibidify:ro \
  -e XDG_CONFIG_HOME=/config \
  ghcr.io/ivuorinen/gibidify:<tag> \
  -source /workspace/your_source_directory \
  -destination /workspace/output.txt \
  --prefix="Your prefix text" \
  --suffix="Your suffix text"
```

## Configuration

gibidify supports a YAML configuration file. Place it at:

- `$XDG_CONFIG_HOME/gibidify/config.yaml` (when `XDG_CONFIG_HOME` is set to an
  absolute path), or
- `$HOME/.config/gibidify/config.yaml`.

Unknown keys are reported as warnings and ignored. A config file that cannot be
read or parsed, or that contains an invalid value, stops gibidify with exit
status 1 and names the problem; it never runs on defaults in that case, because
that would drop the directories you excluded.

Example configuration:

```yaml
fileSizeLimit: 5242880  # 5 MB
ignoreDirectories:
  - vendor
  - node_modules
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
