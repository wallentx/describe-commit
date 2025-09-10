package git

import (
	"context"
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