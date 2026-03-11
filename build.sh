#!/usr/bin/env bash
set -e

LINT_VERSION="${LINT_VERSION:-latest}"
LINT_CMD="golangci-lint run"

ensure_linter() {
    if command -v golangci-lint >/dev/null 2>&1; then
        return
    fi

    echo "golangci-lint not found. Installing ${LINT_VERSION}..."

    if ! command -v go >/dev/null 2>&1; then
        echo "Go is not installed; please install golangci-lint ${LINT_VERSION} manually." >&2
        exit 1
    fi

    GO111MODULE=on go install "github.com/golangci/golangci-lint/cmd/golangci-lint@${LINT_VERSION}"

    # ensure the newly installed binary is on PATH for this session
    local gobin
    gobin="$(go env GOBIN)"
    if [[ -z "${gobin}" ]]; then
        gobin="$(go env GOPATH)/bin"
    fi

    export PATH="${gobin}:${PATH}"
}

ensure_linter

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
