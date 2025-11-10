#!/bin/sh
set -eu

echo "Updating all dependencies to latest versions..."
go get -u ./...
go mod tidy
echo ""
echo "Dependencies updated successfully!"
echo "Running tests to verify compatibility..."
go test ./...
echo ""
echo "Update complete. Run 'make lint-fix && make test' to verify."
