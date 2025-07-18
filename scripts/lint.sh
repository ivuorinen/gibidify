#!/bin/bash
set -e

echo "Running golangci-lint..."
golangci-lint run ./...

echo "Running checkmake..."
checkmake --config=.checkmake Makefile

echo "Running shfmt check..."
shfmt -d .

echo "Running yamllint..."
yamllint -c .yamllint .
