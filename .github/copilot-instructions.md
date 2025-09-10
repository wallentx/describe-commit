# Describe Commit - AI-Powered Git Commit Message Generator

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

`describe-commit` is a Go CLI application that leverages AI to generate commit messages based on Git repository changes. It supports multiple AI providers (OpenAI, Gemini, OpenRouter, Anthropic) and can be built as a standalone binary or Docker image.

### Prerequisites and Setup
- Ensure Go 1.24+ is installed: `go version` (project uses Go 1.24)
- Ensure Git is available: `git --version`
- Repository uses Go modules - no additional dependency management needed

### Bootstrap, Build, and Test the Repository

**CRITICAL**: Never cancel build or test commands. Wait for all operations to complete.

1. **Generate code** (takes ~7 seconds):
   ```bash
   go generate -skip readme ./...
   ```

2. **Build the application** (takes ~15 seconds - NEVER CANCEL):
   ```bash
   go build -o describe-commit ./cmd/describe-commit/
   ```
   - Set timeout to 30+ seconds minimum
   - Binary will be created as `describe-commit` in repository root

3. **Run tests** (takes ~23 seconds - NEVER CANCEL):
   ```bash
   go test -race -covermode=atomic ./...
   ```
   - Set timeout to 45+ seconds minimum  
   - Tests include race detection and coverage analysis
   - All tests should pass with good coverage (>75% for most packages)

4. **Install and run linter** (takes ~8 seconds - NEVER CANCEL):
   ```bash
   # Install golangci-lint if not available
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
   
   # Add to PATH and run linter
   export PATH=$PATH:$(go env GOPATH)/bin
   golangci-lint run
   ```
   - Set timeout to 30+ seconds minimum
   - Should return "0 issues" or display specific linting errors to fix

5. **Build Docker image** (takes ~20 seconds - NEVER CANCEL):
   ```bash
   docker build -t describe-commit:local --build-arg APP_VERSION=local@test .
   ```
   - Set timeout to 60+ seconds minimum
   - Uses multi-stage build (compile stage + runtime Alpine Linux)

### Running the Application

**Test basic functionality:**
```bash
# Show help and validate binary works
./describe-commit --help

# Test with current repository (will fail without API key - expected)
./describe-commit .
# Expected output: "error: invalid options: gemini API key is required"
```

**Docker usage:**
```bash
# Test Docker image
docker run --rm describe-commit:local --help
```

### Validation

**ALWAYS run these validation steps after making changes:**

1. **Build validation**: Ensure the application builds successfully:
   ```bash
   go build -o describe-commit ./cmd/describe-commit/
   ./describe-commit --version
   ```

2. **Test validation**: Run the full test suite:
   ```bash
   go test -race -covermode=atomic ./...
   ```
   Expected result: All tests pass, good coverage (>75% for most packages), shows "ok" for each package

3. **Lint validation**: Ensure code meets quality standards:
   ```bash
   golangci-lint run
   ```
   Expected result: "0 issues" (may show deprecation warnings which are safe to ignore)

4. **Functional validation**: Test core application functionality:
   ```bash
   # Test help command
   ./describe-commit --help
   
   # Test with directory (should show clear error about missing API key)
   ./describe-commit . 2>&1 | grep "API key is required"
   ```

5. **Docker validation**: Ensure Docker build works:
   ```bash
   docker build -t describe-commit:test .
   docker run --rm describe-commit:test --version
   ```

**Manual testing scenarios** (when API keys are available):
- Create staged git changes: `echo "# Test file" > test.txt && git add test.txt`
- Run describe-commit with appropriate API key environment variable:
  ```bash
  # Example with Gemini (free tier)
  GEMINI_API_KEY="your-key" ./describe-commit .
  
  # Example with short message and emoji
  GEMINI_API_KEY="your-key" ./describe-commit -s -e .
  ```
- Clean up: `git reset HEAD test.txt && rm test.txt`
- Expected behavior: AI-generated commit message based on staged changes

### CI/CD Integration

**ALWAYS run these commands before committing** (mirrors CI pipeline):
```bash
# Full CI validation sequence
go generate -skip readme ./...
golangci-lint run
go test -race -covermode=atomic ./...
go build -o describe-commit ./cmd/describe-commit/
```

**GitHub Actions workflows:**
- `.github/workflows/tests.yml`: Runs on push/PR, includes linting, testing, building for multiple platforms
- `.github/workflows/release.yml`: Triggered on releases, builds packages and Docker images

### Key Project Structure

```
.
├── cmd/describe-commit/main.go    # Application entry point
├── internal/                     # Internal packages
│   ├── ai/                      # AI provider implementations
│   ├── cli/                     # CLI command handling
│   ├── config/                  # Configuration management
│   ├── git/                     # Git operations
│   └── version/                 # Version information
├── .golangci.yml                # Linter configuration
├── .goreleaser.yml              # Release configuration
├── Dockerfile                   # Multi-stage Docker build
├── describe-commit.example.yml  # Example configuration file
└── go.mod                       # Go module definition
```

### Configuration

**Configuration file locations** (checked in order):
1. `--config-file` flag or `CONFIG_FILE` environment variable
2. `.describe-commit.yml` or `describe-commit.yml` in working directory or parent directories
3. User config directory:
   - Linux: `~/.config/describe-commit.yml`
   - Windows: `%APPDATA%\describe-commit.yml`
   - macOS: `~/Library/Application Support/describe-commit.yml`

**Environment variables** for testing:
- `GEMINI_API_KEY`: Gemini API key (free tier available)
- `OPENAI_API_KEY`: OpenAI API key (requires paid account)
- `OPENROUTER_API_KEY`: OpenRouter API key
- `ANTHROPIC_API_KEY`: Anthropic API key

### Common Tasks

**Generate README** (updates CLI documentation):
```bash
go generate ./...
```

**Build for specific platform**:
```bash
GOOS=linux GOARCH=amd64 go build -o describe-commit-linux-amd64 ./cmd/describe-commit/
GOOS=darwin GOARCH=arm64 go build -o describe-commit-darwin-arm64 ./cmd/describe-commit/
GOOS=windows GOARCH=amd64 go build -o describe-commit-windows-amd64.exe ./cmd/describe-commit/
```

**Release build** (with version and optimizations):
```bash
VERSION="v1.0.0@$(git rev-parse --short HEAD)"
go build -trimpath -ldflags "-s -w -X gh.tarampamp.am/describe-commit/internal/version.version=${VERSION}" -o describe-commit ./cmd/describe-commit/
```

### Troubleshooting

**Build failures:**
- Ensure Go 1.24+ is installed
- Run `go mod tidy` if dependency issues occur
- Check `go env` for proper GOPATH/GOROOT setup

**Test failures:**
- Use `go test -v ./...` for verbose output
- Individual package testing: `go test -v ./internal/config`
- Race condition debugging: `go test -race -count=100 ./internal/errgroup`

**Linter issues:**
- Check `.golangci.yml` for specific linter configuration
- Use `golangci-lint run --fix` for auto-fixable issues
- Skip false positives with `//nolint:lintername` comments

**Docker build issues:**
- Ensure Docker daemon is running
- Check available disk space for multi-stage builds
- Use `docker system prune` to clean up build cache if needed

### Critical Reminders

- **NEVER CANCEL builds, tests, or linting** - they complete quickly (under 30 seconds each)
- **ALWAYS validate functionality** after changes using the validation scenarios above
- **SET APPROPRIATE TIMEOUTS**: 30s for builds, 45s for tests, 60s for Docker builds
- **API keys required** for functional testing - application will clearly indicate missing keys
- **Configuration files** use YAML format - see `describe-commit.example.yml` for reference