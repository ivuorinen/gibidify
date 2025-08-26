#!/usr/bin/env bash

echo "Installing revive..."
go install github.com/mgechev/revive@v1.11.0
echo "Installing gofumpt..."
go install mvdan.cc/gofumpt@v0.8.0
echo "Installing goimports..."
go install golang.org/x/tools/cmd/goimports@v0.36.0
echo "Installing staticcheck..."
go install honnef.co/go/tools/cmd/staticcheck@v0.6.1
echo "Installing gosec..."
go install github.com/securego/gosec/v2/cmd/gosec@v2.22.8
echo "Installing gocyclo..."
go install github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0
echo "Installing checkmake..."
go install github.com/checkmake/checkmake/cmd/checkmake@0.2.2
echo "Installing shfmt..."
go install mvdan.cc/sh/v3/cmd/shfmt@v3.12.0
echo "Installing yaml-lint (mvdan.cc)…"
go install mvdan.cc/yaml/cmd/yaml-lint@v2.4.0
echo "Installing yamlfmt (Google)…"
go install github.com/google/yamlfmt/cmd/yamlfmt@v0.4.0
echo "Installing eclint..."
go install gitlab.com/greut/eclint/cmd/eclint@v0.5.1
echo "All tools installed successfully!"
