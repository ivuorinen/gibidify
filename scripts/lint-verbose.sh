#!/bin/sh
set -eu

echo "Running golangci-lint (verbose)..."
golangci-lint run -v ./...
echo "Running checkmake (verbose)..."
checkmake --config=.checkmake --format="{{.Line}}:{{.Rule}}:{{.Violation}}" Makefile
echo "Running shfmt check (verbose)..."
shfmt -d .
echo "Running yamllint (verbose)..."
yamllint .
