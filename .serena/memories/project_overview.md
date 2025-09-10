# Gibidify Project Overview

## Purpose
Gibidify is a Go CLI application that scans source directories recursively and aggregates code files into a single text file optimized for Large Language Models (LLMs). It supports multiple output formats (markdown, JSON, YAML) with intelligent file filtering and concurrent processing.

## Tech Stack
- **Language**: Go 1.25.0
- **Key Dependencies**:
  - CLI: fatih/color, schollz/progressbar/v3
  - Config: spf13/viper, gopkg.in/yaml.v3
  - File handling: sabhiram/go-gitignore
  - Text processing: golang.org/x/text
  - Logging: sirupsen/logrus

## Architecture
- **Core modules**: main.go, cli/, fileproc/, config/, shared/, testutil/
- **Advanced features**: metrics/, templates/, benchmark/
- **Key patterns**: Producer-consumer, thread-safe registry, streaming, modular design (50-200 lines per module)
- **Performance**: ~63ns cache registry lookups, memory-optimized processing

## Current Status
- **Health**: 9/10 - Production-ready
- **Test Coverage**: 77.9% overall
- **Lines of Code**: ~21.5K across 92 files
- **Branch**: fix/tests-and-linting (main branch: main)
