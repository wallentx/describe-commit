#!/usr/bin/env bash
set -e

LINT_CMD="go tool golangci-lint run"

# Check for -f flag
if [[ "$1" == "-f" ]]; then
    echo "Fix flag detected. Running linter with --fix..."
    LINT_CMD="$LINT_CMD --fix"
else
    echo "Running linter..."
fi

# Run linter
$LINT_CMD
echo "Linting passed."

# Build binary
echo "Building describe-commit binary..."
go build -o describe-commit ./cmd/describe-commit/
echo "Build complete: ./describe-commit"
