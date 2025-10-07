#!/bin/sh
set -eu

echo "Running golangci-lint..."
golangci-lint run ./...

echo "Running revive..."
revive -config revive.toml -formatter friendly ./...

echo "Running checkmake..."
checkmake --config=.checkmake Makefile

echo "Running shfmt check..."
shfmt -d -i 2 -ci .

echo "Running yamllint..."
yamllint .
