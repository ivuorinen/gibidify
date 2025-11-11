#!/bin/sh
set -eu

echo "Checking for available dependency updates..."
echo ""
echo "Direct dependencies:"
go list -u -m all | grep -v "indirect" | column -t
echo ""
echo "Note: Run 'make deps-update' to update all dependencies"
