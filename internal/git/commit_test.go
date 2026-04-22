package git

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCommitDiff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Use current directory as test repository
	ctx := context.Background()

	// Test with an invalid commit hash - should return error
	_, err := CommitDiff(ctx, ".", "invalid-commit-hash")
	if err == nil {
		t.Error("expected error for invalid commit hash, but got none")
	}

	// Test with empty directory path - should return error
	_, err = CommitDiff(ctx, "", "abc123")
	if err == nil {
		t.Error("expected error for empty directory path, but got none")
	}
}

func TestCommitLog(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Use current directory as test repository
	ctx := context.Background()

	// Test with an invalid commit hash - should return error
	_, err := CommitLog(ctx, ".", "invalid-commit-hash", 5)
	if err == nil {
		t.Error("expected error for invalid commit hash, but got none")
	}

	// Test with empty directory path - should return error
	_, err = CommitLog(ctx, "", "abc123", 5)
	if err == nil {
		t.Error("expected error for empty directory path, but got none")
	}
}

func TestCommitLogWithRootCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "user.email", "test@example.com")

	filePath := dir + "/README.md"
	if err := os.WriteFile(filePath, []byte("hello\n"), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "initial commit", "--quiet", "--no-gpg-sign")
	commitHash := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))

	logOutput, err := CommitLog(context.Background(), dir, commitHash, 5)
	if err != nil {
		t.Fatalf("expected no error for root commit history, got: %v", err)
	}

	if logOutput != "" {
		t.Fatalf("expected no prior history for root commit, got: %q", logOutput)
	}
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"LC_ALL=C", "LANG=C",
		"NO_COLOR=1",
		"GIT_CONFIG_NOSYSTEM=1",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v, output: %s", args, err, out)
	}

	return string(out)
}
