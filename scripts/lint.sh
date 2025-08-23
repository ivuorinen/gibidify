#!/bin/bash

# Track overall exit status
exit_code=0

echo "Running golangci-lint..."
if ! golangci-lint run ./...; then
  exit_code=1
fi

echo "Running checkmake..."
if ! checkmake --config=.checkmake Makefile; then
  exit_code=1
fi

echo "Running shfmt check..."
if ! shfmt -d .; then
  exit_code=1
fi

echo "Running yamllint..."
if ! yamllint -c .yamllint .; then
  exit_code=1
fi

echo "Running eclint..."
if ! eclint; then
  exit_code=1
fi

# Exit with failure status if any linter failed
exit $exit_code
