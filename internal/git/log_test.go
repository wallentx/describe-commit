package git

import (
	"context"
	"os"
	"os/exec"
	"testing"
)

func TestLogWithEmptyRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	dir := t.TempDir()

	initCmd := exec.Command("git", "init")
	initCmd.Dir = dir
	initCmd.Env = append(os.Environ(),
		"LC_ALL=C", "LANG=C",
		"NO_COLOR=1",
		"GIT_CONFIG_NOSYSTEM=1",
	)

	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v, output: %s", err, out)
	}

	logOutput, err := Log(context.Background(), dir, 5)
	if err != nil {
		t.Fatalf("expected no error for repository without commits, got: %v", err)
	}

	if logOutput != "" {
		t.Fatalf("expected empty log output for repository without commits, got: %q", logOutput)
	}
}
