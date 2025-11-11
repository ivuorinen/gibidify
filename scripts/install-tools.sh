#!/bin/sh
set -eu

echo "Installing golangci-lint..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
echo "Installing gofumpt..."
go install mvdan.cc/gofumpt@latest
echo "Installing golines..."
go install github.com/segmentio/golines@latest
echo "Installing goimports..."
go install golang.org/x/tools/cmd/goimports@latest
echo "Installing staticcheck..."
go install honnef.co/go/tools/cmd/staticcheck@latest
echo "Installing gosec..."
go install github.com/securego/gosec/v2/cmd/gosec@latest
echo "Installing gocyclo..."
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
echo "Installing revive..."
go install github.com/mgechev/revive@latest
echo "Installing checkmake..."
go install github.com/checkmake/checkmake/cmd/checkmake@latest
echo "Installing shellcheck..."
go install github.com/koalaman/shellcheck/cmd/shellcheck@latest
echo "Installing shfmt..."
go install mvdan.cc/sh/v3/cmd/shfmt@latest
echo "Installing yamllint (Go-based)..."
go install github.com/excilsploft/yamllint@latest
echo "Installing editorconfig-checker..."
go install github.com/editorconfig-checker/editorconfig-checker/cmd/editorconfig-checker@latest
echo "All tools installed successfully!"
