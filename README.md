# gibidi

gibidi is a CLI application written in Go that scans a source directory recursively and aggregates code files into a single text file. The output file is designed for use with LLM models and includes a prefix, file sections with separators, and a suffix.

## Features

- Recursive scanning of a source directory.
- File filtering based on size, glob patterns, and .gitignore rules.
- Modular, concurrent file processing with progress bar feedback.
- Configurable logging and configuration via Viper.
- Cross-platform build with Docker packaging support.

## Installation

Clone the repository and build the application:

```bash
git clone https://github.com/ivuorinen/gibidi.git
cd gibidify
go build -o gibidify ./cmd/gibidify
```

## Usage

```bash
./gibidify -source <source_directory> -destination <output_file> [--prefix="..."] [--suffix="..."]
```

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

gibidi supports a YAML configuration file. Place it at:

- `$XDG_CONFIG_HOME/gibidi/config.yaml` or
- `$HOME/.config/gibidi/config.yaml` or
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
  - bower_components
  - cache
  - tmp
```

## License

This project is licensed under the MIT License.
