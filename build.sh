#!/usr/bin/env bash
set -e

if go tool -n golangci-lint >/dev/null 2>&1; then
    LINT_CMD=(go tool golangci-lint run ./...)
elif command -v golangci-lint >/dev/null 2>&1; then
    LINT_CMD=(golangci-lint run ./...)
else
    echo "golangci-lint not found in go tool or PATH" >&2
    exit 1
fi

# Check for -f flag
if [[ "$1" == "-f" ]]; then
    echo "Fix flag detected. Running linter with --fix..."
    LINT_CMD+=(--fix)
else
    echo "Running linter..."
fi

# Run linter
"${LINT_CMD[@]}"
echo "Linting passed."

# Build binary
echo "Building describe-commit binary..."
go build -o describe-commit ./cmd/describe-commit/
echo "Build complete: ./describe-commit"
