#!/bin/sh
set -eu

echo "Running tests with coverage..."
go test -race -v -coverprofile=coverage.out -covermode=atomic ./...
echo ""
echo "Coverage summary:"
go tool cover -func=coverage.out | grep total:
echo ""
echo "Full coverage report saved to: coverage.out"
echo "To view HTML report, run: make coverage"
