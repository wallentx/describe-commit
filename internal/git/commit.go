package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CommitDiff returns the diff for a specific commit.
func CommitDiff(ctx context.Context, dirPath string, commitHash string) (string, error) {
	// ensure git is installed and available to run
	gitFilePath, lookErr := binPath()
	if lookErr != nil {
		return "", lookErr
	}

	// get the diff for the specific commit
	var cmd = exec.CommandContext(ctx, gitFilePath, "show",
		"--format=",                // don't show commit message
		"--ignore-submodules=all",  // ignore changes to submodules
		"--diff-algorithm=minimal", // use the minimal diff algorithm
		"--no-ext-diff",            // do not use external diff helper
		"--ignore-all-space",       // ignore whitespace when comparing lines
		"--ignore-blank-lines",     // ignore changes whose lines are all blank
		"--no-color",               // do not use any color in the output
		"--patch",                  // generate patch (unified diff) format
		commitHash,
		"--",
		":(exclude)*.sum",  // exclude .sum files
		":(exclude)*.lock", // exclude .lock files
		":(exclude)*.log",  // exclude .log files
		":(exclude)*.out",  // exclude .out files
		":(exclude)*.tmp",  // exclude .tmp files
		":(exclude)*.bak",  // exclude .bak files
		":(exclude)*.swp",  // exclude .swp files
		":(exclude)*.env",  // exclude .env files
	)

	cmd.Dir = dirPath
	cmd.Env = []string{
		"LC_ALL=C", "LANG=C", // forces the system to use the "C" (POSIX) locale, English-based output with no localization
		"NO_COLOR=1",            // disables colored output
		"GIT_CONFIG_NOSYSTEM=1", // do not use the system-wide configuration file
	}

	var stdOut, stdErr bytes.Buffer

	stdOut.Grow(1024 * 8) //nolint:mnd // 8KB

	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		if stdErr.Len() > 0 {
			err = fmt.Errorf("%s: %w", stdErrToString(stdErr.String()), err)
		}

		return "", fmt.Errorf("git show failed for commit %s: %w", commitHash, err)
	}

	return stdOut.String(), nil
}

// CommitLog returns the commit log of the repository starting from a specific commit, limited to the specified number of commits.
func CommitLog(ctx context.Context, dirPath string, commitHash string, len int) (string, error) {
	// ensure git is installed and available to run
	gitFilePath, lookErr := binPath()
	if lookErr != nil {
		return "", lookErr
	}

	parentHash, parentErr := firstParent(ctx, gitFilePath, dirPath, commitHash)
	if parentErr != nil {
		return "", parentErr
	}

	if parentHash == "" {
		return "", nil // no history exists before the root commit
	}

	// get the log starting from the commit's parent (to get context)
	var cmd = exec.CommandContext(ctx, gitFilePath, "log",
		"--format=%s",
		fmt.Sprintf("--max-count=%d", len),
		"--no-color",
		parentHash,
	)

	cmd.Dir = dirPath
	cmd.Env = []string{
		"LC_ALL=C", "LANG=C", // forces the system to use the "C" (POSIX) locale, English-based output with no localization
		"NO_COLOR=1",            // disables colored output
		"GIT_CONFIG_NOSYSTEM=1", // do not use the system-wide configuration file
	}

	var stdOut, stdErr bytes.Buffer

	stdOut.Grow(1024 * 2) //nolint:mnd // 2KB

	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		if stdErr.Len() > 0 {
			err = fmt.Errorf("%s: %w", stdErrToString(stdErr.String()), err)
		}

		return "", fmt.Errorf("git log failed for commit %s: %w", commitHash, err)
	}

	return stdOut.String(), nil
}

func firstParent(ctx context.Context, gitFilePath, dirPath, commitHash string) (string, error) {
	cmd := exec.CommandContext(ctx, gitFilePath, "rev-list", "--parents", "-n", "1", commitHash)
	cmd.Dir = dirPath
	cmd.Env = []string{
		"LC_ALL=C", "LANG=C",
		"NO_COLOR=1",
		"GIT_CONFIG_NOSYSTEM=1",
	}

	var stdOut, stdErr bytes.Buffer

	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		if stdErr.Len() > 0 {
			err = fmt.Errorf("%s: %w", stdErrToString(stdErr.String()), err)
		}

		return "", fmt.Errorf("git rev-list failed for commit %s: %w", commitHash, err)
	}

	fields := strings.Fields(stdOut.String())
	if len(fields) == 0 {
		return "", fmt.Errorf("git rev-list returned no commits for %s", commitHash)
	}

	if len(fields) == 1 {
		return "", nil
	}

	return fields[1], nil
}
