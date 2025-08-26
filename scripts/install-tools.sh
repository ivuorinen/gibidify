#!/usr/bin/env bash

echo "Installing revive..."
go install github.com/mgechev/revive@latest
echo "Installing gofumpt..."
go install mvdan.cc/gofumpt@latest
echo "Installing goimports..."
go install golang.org/x/tools/cmd/goimports@latest
echo "Installing staticcheck..."
go install honnef.co/go/tools/cmd/staticcheck@latest
echo "Installing gosec..."
go install github.com/securego/gosec/v2/cmd/gosec@latest
echo "Installing gocyclo..."
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
echo "Installing checkmake..."
go install github.com/checkmake/checkmake/cmd/checkmake@latest
echo "Installing shfmt..."
go install mvdan.cc/sh/v3/cmd/shfmt@latest
echo "Installing yamllint (Go-based)..."
go install github.com/excilsploft/yamllint@latest
echo "Installing eclint..."
go install gitlab.com/greut/eclint/cmd/eclint@latest
echo "All tools installed successfully!"
