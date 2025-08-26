echo "Running gofumpt..."
gofumpt -l -w .
echo "Running goimports..."
goimports -w -local github.com/ivuorinen/gibidify .
echo "Running go fmt..."
go fmt ./...
echo "Running go mod tidy..."
go mod tidy
echo "Running shfmt formatting..."
shfmt -w -i 2 -ci .
echo "Running revive linter..."
revive -config revive.toml -formatter friendly -set_exit_status ./...
echo "Running gosec security linter..."
gosec -fmt=text -quiet ./...
echo "Auto-fix completed. Running final lint check..."
revive -config revive.toml -formatter friendly -set_exit_status ./...
gosec -fmt=text -quiet ./...
echo "Running checkmake..."
checkmake --config=.checkmake Makefile
echo "Running yamlfmt..."
yamlfmt -conf .yamlfmt.yml -gitignore_excludes -dstar ./**/*.{yaml,yml}
echo "Running eclint fix..."
eclint -fix
