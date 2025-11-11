#!/bin/sh
set -eu

echo "Running gofumpt..."
gofumpt -l -w .
echo "Running golines..."
golines -w -m 120 --base-formatter="gofumpt" --shorten-comments .
echo "Running goimports..."
goimports -w -local github.com/ivuorinen/gibidify .
echo "Running go fmt..."
go fmt ./...
echo "Running go mod tidy..."
go mod tidy
echo "Running shfmt formatting..."
shfmt -w -i 0 -ci .
echo "Running golangci-lint with --fix..."
golangci-lint run --fix ./...
echo "Auto-fix completed. Running final lint check..."
golangci-lint run ./...
echo "Running revive..."
revive -config revive.toml -formatter friendly ./...
echo "Running checkmake..."
checkmake --config=.checkmake Makefile
echo "Running yamllint..."
yamllint .
