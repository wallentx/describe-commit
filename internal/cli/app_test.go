package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gh.tarampamp.am/describe-commit/internal/ai"
	"gh.tarampamp.am/describe-commit/internal/cli/cmd"
)

func TestSetIfFlagIsSet_UsesExplicitValueMatchingDefault(t *testing.T) {
	target := true
	value := false

	setIfFlagIsSet(&target, cmd.Flag[bool]{
		Default:      false,
		Value:        &value,
		ValueSetFrom: cmd.FlagValueSourceFlag,
	})

	if target {
		t.Fatal("expected explicit flag value to override target even when it matches the default")
	}
}

func TestLoadOptions_DoesNotLeakCallerConfigIntoTargetRepo(t *testing.T) {
	baseDir := t.TempDir()
	callerDir := filepath.Join(baseDir, "caller")
	targetDir := filepath.Join(baseDir, "target")

	for _, dir := range []string{callerDir, targetDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}
	}

	writeConfigFile(t, filepath.Join(callerDir, "describe-commit.yml"), `
shortMessageOnly: true
aiProvider: openai
`)
	writeConfigFile(t, filepath.Join(targetDir, "describe-commit.yml"), `
gemini:
  apiKey: target-key
  modelName: target-model
`)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err = os.Chdir(callerDir); err != nil {
		t.Fatalf("failed to switch working directory: %v", err)
	}

	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("failed to restore working directory: %v", chdirErr)
		}
	})

	app := NewApp("describe-commit")

	resolvedWD, loadErr := app.loadOptions(context.Background(), "", []string{targetDir}, func() {})
	if loadErr != nil {
		t.Fatalf("failed to load options: %v", loadErr)
	}

	if resolvedWD != targetDir {
		t.Fatalf("unexpected working directory: %s", resolvedWD)
	}

	if app.opt.ShortMessageOnly {
		t.Fatal("expected target repo to use defaults when its config does not set shortMessageOnly")
	}

	if app.opt.AIProviderName != ai.ProviderGemini {
		t.Fatalf("unexpected AI provider: %s", app.opt.AIProviderName)
	}
}

func writeConfigFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write config file %s: %v", path, err)
	}
}
