package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gh.tarampamp.am/describe-commit/internal/ai"
	"gh.tarampamp.am/describe-commit/internal/cli/cmd"
	"gh.tarampamp.am/describe-commit/internal/config"
	"gh.tarampamp.am/describe-commit/internal/debug"
	"gh.tarampamp.am/describe-commit/internal/errgroup"
	"gh.tarampamp.am/describe-commit/internal/git"
	"gh.tarampamp.am/describe-commit/internal/version"
)

//go:generate go run ./generate/readme.go

type App struct {
	cmd cmd.Command
	opt options
}

func NewApp(name string) *App { //nolint:funlen
	var app = App{
		cmd: cmd.Command{
			Name:        name,
			Description: "This tool leverages AI to generate commit messages based on changes made in a Git repository.",
			Usage:       "[<options>] [<git-dir-path>]",
			Version:     version.Version(),
		},
		opt: newOptionsWithDefaults(),
	}

	var (
		configFile = cmd.Flag[string]{
			Names:   []string{"config-file", "c"},
			Usage:   "Path to the configuration file",
			EnvVars: []string{"CONFIG_FILE"},
			Default: filepath.Join(config.DefaultDirPath(), config.FileName),
		}
		shortMessageOnly = cmd.Flag[bool]{
			Names:   []string{"short-message-only", "s"},
			Usage:   "Generate a short commit message (subject line) only",
			EnvVars: []string{"SHORT_MESSAGE_ONLY"},
			Default: app.opt.ShortMessageOnly,
		}
		commitHistoryLength = cmd.Flag[int64]{
			Names:   []string{"commit-history-length", "cl", "hl"},
			Usage:   "Number of previous commits from the Git history (0 = disabled)",
			EnvVars: []string{"COMMIT_HISTORY_LENGTH"},
			Default: app.opt.CommitHistoryLength,
		}
		enableEmoji = cmd.Flag[bool]{
			Names:   []string{"enable-emoji", "e"},
			Usage:   "Enable emoji in the commit message",
			EnvVars: []string{"ENABLE_EMOJI"},
			Default: app.opt.EnableEmoji,
		}
		maxOutputTokens = cmd.Flag[int64]{
			Names:   []string{"max-output-tokens"},
			Usage:   "Maximum number of tokens in the output message",
			EnvVars: []string{"MAX_OUTPUT_TOKENS"},
			Validator: func(_ *cmd.Command, i int64) error {
				if i <= 1 {
					return errors.New("max output tokens must be greater than 1")
				}

				return nil
			},
			Default: app.opt.MaxOutputTokens,
		}
		aiProviderName = cmd.Flag[string]{
			Names:   []string{"ai-provider", "ai"},
			Usage:   fmt.Sprintf("AI provider name (%s)", strings.Join(ai.SupportedProviders(), "|")),
			EnvVars: []string{"AI_PROVIDER"},
			Default: app.opt.AIProviderName,
			Validator: func(_ *cmd.Command, s string) error {
				if !ai.IsProviderSupported(s) {
					return fmt.Errorf("unsupported AI provider: %s", s)
				}

				return nil
			},
		}
		geminiApiKey = cmd.Flag[string]{
			Names:   []string{"gemini-api-key", "ga"},
			Usage:   "Gemini API key (https://bit.ly/4jZhiKI, as of February 2025 it's free)",
			EnvVars: []string{"GEMINI_API_KEY"},
			Default: app.opt.Providers.Gemini.ApiKey,
		}
		geminiModelName = cmd.Flag[string]{
			Names:   []string{"gemini-model-name", "gm"},
			Usage:   "Gemini model name (https://bit.ly/4i02ARR)",
			EnvVars: []string{"GEMINI_MODEL_NAME"},
			Default: app.opt.Providers.Gemini.ModelName,
		}
		openAIApiKey = cmd.Flag[string]{
			Names:   []string{"openai-api-key", "oa"},
			Usage:   "OpenAI API key (https://bit.ly/4i03NbR, you need to add funds to your account)",
			EnvVars: []string{"OPENAI_API_KEY"},
			Default: app.opt.Providers.OpenAI.ApiKey,
		}
		openAIModelName = cmd.Flag[string]{
			Names:   []string{"openai-model-name", "om"},
			Usage:   "OpenAI model name (https://bit.ly/4hXCXkL)",
			EnvVars: []string{"OPENAI_MODEL_NAME"},
			Default: app.opt.Providers.OpenAI.ModelName,
		}
		openRouterApiKey = cmd.Flag[string]{
			Names:   []string{"openrouter-api-key", "ora"},
			Usage:   "OpenRouter API key (https://bit.ly/4hU1yY1)",
			EnvVars: []string{"OPENROUTER_API_KEY"},
			Default: app.opt.Providers.OpenRouter.ApiKey,
		}
		openRouterModelName = cmd.Flag[string]{
			Names:   []string{"openrouter-model-name", "orm"},
			Usage:   "OpenRouter model name (https://bit.ly/4ktktuG)",
			EnvVars: []string{"OPENROUTER_MODEL_NAME"},
			Default: app.opt.Providers.OpenRouter.ModelName,
		}
		anthropicApiKey = cmd.Flag[string]{
			Names:   []string{"anthropic-api-key", "ana"},
			Usage:   "Anthropic API key (https://bit.ly/4klw0Mw)",
			EnvVars: []string{"ANTHROPIC_API_KEY"},
			Default: app.opt.Providers.Anthropic.ApiKey,
		}
		anthropicModelName = cmd.Flag[string]{
			Names:   []string{"anthropic-model-name", "anm"},
			Usage:   "Anthropic model name (https://bit.ly/4bmQDDV)",
			EnvVars: []string{"ANTHROPIC_MODEL_NAME"},
			Default: app.opt.Providers.Anthropic.ModelName,
		}
		commitHash = cmd.Flag[string]{
			Names:   []string{"commit-hash", "ch"},
			Usage:   "Specific commit hash to describe (instead of staged changes)",
			EnvVars: []string{"COMMIT_HASH"},
			Default: app.opt.CommitHash,
		}
		repoURL = cmd.Flag[string]{
			Names:   []string{"repo", "r"},
			Usage:   "Repository URL in format 'owner/reponame' (requires --commit-hash)",
			EnvVars: []string{"REPO_URL"},
			Default: app.opt.RepoURL,
		}
		branch = cmd.Flag[string]{
			Names:   []string{"branch", "b"},
			Usage:   "Git branch to use when cloning repository (used with --repo)",
			EnvVars: []string{"BRANCH"},
			Default: app.opt.Branch,
		}
	)

	app.cmd.Flags = []cmd.Flagger{
		&configFile,
		&shortMessageOnly,
		&commitHistoryLength,
		&enableEmoji,
		&maxOutputTokens,
		&aiProviderName,
		&geminiApiKey,
		&geminiModelName,
		&openAIApiKey,
		&openAIModelName,
		&openRouterApiKey,
		&openRouterModelName,
		&anthropicApiKey,
		&anthropicModelName,
		&commitHash,
		&repoURL,
		&branch,
	}

	app.cmd.Action = func(ctx context.Context, c *cmd.Command, args []string) error {
		// update the options from the configuration file(s)
		if err := app.opt.UpdateFromConfigFile(append([]string{*configFile.Value}, config.FindIn(".")...)); err != nil {
			return err
		}

		{ // override the options with the command-line flags
			setIfFlagIsSet(&app.opt.ShortMessageOnly, shortMessageOnly)
			setIfFlagIsSet(&app.opt.CommitHistoryLength, commitHistoryLength)
			setIfFlagIsSet(&app.opt.EnableEmoji, enableEmoji)
			setIfFlagIsSet(&app.opt.MaxOutputTokens, maxOutputTokens)
			setIfFlagIsSet(&app.opt.AIProviderName, aiProviderName)
			setIfFlagIsSet(&app.opt.Providers.Gemini.ApiKey, geminiApiKey)
			setIfFlagIsSet(&app.opt.Providers.Gemini.ModelName, geminiModelName)
			setIfFlagIsSet(&app.opt.Providers.OpenAI.ApiKey, openAIApiKey)
			setIfFlagIsSet(&app.opt.Providers.OpenAI.ModelName, openAIModelName)
			setIfFlagIsSet(&app.opt.Providers.OpenRouter.ApiKey, openRouterApiKey)
			setIfFlagIsSet(&app.opt.Providers.OpenRouter.ModelName, openRouterModelName)
			setIfFlagIsSet(&app.opt.Providers.Anthropic.ApiKey, anthropicApiKey)
			setIfFlagIsSet(&app.opt.Providers.Anthropic.ModelName, anthropicModelName)
			setIfFlagIsSet(&app.opt.CommitHash, commitHash)
			setIfFlagIsSet(&app.opt.RepoURL, repoURL)
			setIfFlagIsSet(&app.opt.Branch, branch)
		}

		if err := app.opt.Validate(); err != nil {
			return fmt.Errorf("invalid options: %w", err)
		}

		// determine the working directory
		var wd, wdErr = app.getWorkingDir(ctx, args, app.opt.RepoURL, app.opt.Branch)
		if wdErr != nil {
			return fmt.Errorf("wrong working directory: %w", wdErr)
		}

		// if repo URL was provided, we need to use a different config file search path
		var configSearchDir = "."
		if app.opt.RepoURL == "" {
			configSearchDir = wd
		}

		// update the options from the configuration file(s) after we know the working directory
		configFiles := append([]string{*configFile.Value}, config.FindIn(configSearchDir)...)
		if err := app.opt.UpdateFromConfigFile(configFiles); err != nil {
			return err
		}

		return app.run(ctx, wd, app.opt.CommitHash)
	}

	return &app
}

// setIfFlagIsSet sets the value from the flag to the option if the flag is set and the value is not nil.
func setIfFlagIsSet[T cmd.FlagType](target *T, source cmd.Flag[T]) {
	if target == nil || source.Value == nil || !source.IsSet() {
		return
	}

	*target = *source.Value
}

// cloneRepoToTemp clones a repository to a temporary directory with minimal checkout.
func cloneRepoToTemp(ctx context.Context, repoURL string, branch string) (string, error) {
	// create a temporary directory
	tempDir, err := os.MkdirTemp("", "describe-commit-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// convert owner/repo format to full GitHub URL if needed
	var fullURL string
	if strings.Contains(repoURL, "/") && !strings.HasPrefix(repoURL, "http") {
		fullURL = "https://github.com/" + repoURL + ".git"
	} else {
		fullURL = repoURL
	}

	// prepare git clone command arguments
	args := []string{"clone",
		"--depth", "50", // get enough commits for history
		"--no-checkout", // don't checkout working tree
	}

	// add branch specification if provided
	if branch != "" {
		args = append(args, "--branch", branch)
	}

	args = append(args, fullURL, tempDir)

	// clone the repository with minimal settings
	cmd := exec.CommandContext(ctx, "git", args...)

	if err := cmd.Run(); err != nil {
		// clean up the temp directory on failure
		_ = os.RemoveAll(tempDir)

		return "", fmt.Errorf("failed to clone repository %s: %w", repoURL, err)
	}

	if branch != "" {
		debug.Printf("cloned repository %s (branch: %s) to temporary directory: %s", repoURL, branch, tempDir)
	} else {
		debug.Printf("cloned repository %s to temporary directory: %s", repoURL, tempDir)
	}

	return tempDir, nil
}

// ensureCommitAvailable ensures a specific commit is available in the repository.
// If the commit is not found, it attempts to fetch it from the remote.
func ensureCommitAvailable(ctx context.Context, dirPath string, commitHash string) error {
	// first check if the commit already exists
	checkCmd := exec.CommandContext(ctx, "git", "cat-file", "-e", commitHash)
	checkCmd.Dir = dirPath
	checkCmd.Env = []string{
		"LC_ALL=C", "LANG=C",
		"NO_COLOR=1",
		"GIT_CONFIG_NOSYSTEM=1",
	}

	if err := checkCmd.Run(); err == nil {
		// commit already exists
		debug.Printf("commit %s is available in repository", commitHash)

		return nil
	}

	debug.Printf("commit %s not found, attempting to fetch from remote", commitHash)

	// try to fetch the specific commit from origin
	fetchCmd := exec.CommandContext(ctx, "git", "fetch", "origin", commitHash)
	fetchCmd.Dir = dirPath
	fetchCmd.Env = []string{
		"LC_ALL=C", "LANG=C",
		"NO_COLOR=1",
		"GIT_CONFIG_NOSYSTEM=1",
	}

	if err := fetchCmd.Run(); err != nil {
		debug.Printf("failed to fetch commit %s from origin: %v", commitHash, err)

		// if that fails, try unshallowing the repository to get more history
		unshallowCmd := exec.CommandContext(ctx, "git", "fetch", "--unshallow")
		unshallowCmd.Dir = dirPath
		unshallowCmd.Env = []string{
			"LC_ALL=C", "LANG=C",
			"NO_COLOR=1",
			"GIT_CONFIG_NOSYSTEM=1",
		}

		if unshallowErr := unshallowCmd.Run(); unshallowErr != nil {
			return fmt.Errorf("commit %s not found and could not fetch from remote: %w", commitHash, err)
		}

		debug.Printf("unshallowed repository to fetch more history")
	} else {
		debug.Printf("successfully fetched commit %s from origin", commitHash)
	}

	// verify the commit is now available
	verifyCmd := exec.CommandContext(ctx, "git", "cat-file", "-e", commitHash)
	verifyCmd.Dir = dirPath
	verifyCmd.Env = []string{
		"LC_ALL=C", "LANG=C",
		"NO_COLOR=1",
		"GIT_CONFIG_NOSYSTEM=1",
	}

	if err := verifyCmd.Run(); err != nil {
		return fmt.Errorf("commit %s still not available after fetch attempts", commitHash)
	}

	debug.Printf("commit %s is now available in repository", commitHash)

	return nil
}

// getWorkingDir returns the working directory to use for the application.
func (*App) getWorkingDir(ctx context.Context, args []string, repoURL string, branch string) (string, error) {
	// if repo URL is provided, clone it to a temporary directory
	if repoURL != "" {
		return cloneRepoToTemp(ctx, repoURL, branch)
	}

	var dir string

	if len(args) > 0 {
		dir = filepath.Clean(strings.TrimSpace(args[0]))
	}

	// if the argument was not set, use the `os.Getwd`
	if dir == "" {
		if d, err := os.Getwd(); err != nil {
			return "", err
		} else {
			dir = d
		}
	}

	// convert the relative path to the absolute one
	if !filepath.IsAbs(dir) {
		if abs, absErr := filepath.Abs(dir); absErr != nil {
			return "", absErr
		} else {
			dir = abs
		}
	}

	// check the working directory existence
	if stat, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("working directory does not exist: %s", dir)
		}

		return "", err
	} else if !stat.IsDir() {
		return "", fmt.Errorf("not a directory: %s", dir)
	}

	return dir, nil
}

// Run runs the application.
func (a *App) Run(ctx context.Context, args []string) error { return a.cmd.Run(ctx, args) }

// Help returns the help message.
func (a *App) Help() string { return a.cmd.Help() }

func providerFromOptions(opt options) (ai.Provider, error) {
	switch opt.AIProviderName {
	case ai.ProviderGemini:
		return ai.NewGemini(
			opt.Providers.Gemini.ApiKey,
			opt.Providers.Gemini.ModelName,
		), nil
	case ai.ProviderOpenAI:
		return ai.NewOpenAI(
			opt.Providers.OpenAI.ApiKey,
			opt.Providers.OpenAI.ModelName,
		), nil
	case ai.ProviderOpenRouter:
		return ai.NewOpenRouter(
			opt.Providers.OpenRouter.ApiKey,
			opt.Providers.OpenRouter.ModelName,
		), nil
	case ai.ProviderAnthropic:
		return ai.NewAnthropic(
			opt.Providers.Anthropic.ApiKey,
			opt.Providers.Anthropic.ModelName,
		), nil
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", opt.AIProviderName)
	}
}

func markMissingHistory(commits string, histLen int64) string {
	if commits == "" && histLen > 0 {
		debug.Printf("git log output is empty: repository has no commits yet")

		return "NO HISTORY AVAILABLE (repository has no commits yet)"
	}

	return commits
}

// run in the main logic of the application.
func (a *App) run(ctx context.Context, workingDir string, commitHash string) error { //nolint:funlen
	debug.Printf("AI provider: %s", a.opt.AIProviderName)

	provider, err := providerFromOptions(a.opt)
	if err != nil {
		return err
	}

	debug.Printf("working directory: %s", workingDir)

	var (
		eg, _            = errgroup.New(ctx)
		changes, commits string
	)

	if commitHash != "" {
		// If we're using a remote repository, ensure the commit is available
		if a.opt.RepoURL != "" {
			if err := ensureCommitAvailable(ctx, workingDir, commitHash); err != nil {
				return fmt.Errorf("failed to ensure commit %s is available: %w", commitHash, err)
			}
		}

		// Get changes for a specific commit
		eg.Go(func(ctx context.Context) (err error) {
			changes, err = git.CommitDiff(ctx, workingDir, commitHash)

			return
		})
	} else {
		// Get staged changes (existing behavior)
		eg.Go(func(ctx context.Context) (err error) {
			changes, err = git.Diff(ctx, workingDir)

			return
		})
	}

	if histLen := int(a.opt.CommitHistoryLength); histLen > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if commitHash != "" {
				commits, err = git.CommitLog(ctx, workingDir, commitHash, histLen)
			} else {
				commits, err = git.Log(ctx, workingDir, histLen)
			}

			return
		})
	} else {
		commits = "NO COMMITS"
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	commits = markMissingHistory(commits, a.opt.CommitHistoryLength)

	debug.Printf("changes:\n%s", changes)
	debug.Printf("commits:\n%s", commits)

	if changes == "" {
		if commitHash != "" {
			return fmt.Errorf("no changes found for commit %s in %s", commitHash, workingDir)
		}

		return fmt.Errorf("no changes found in %s (probably nothing staged; try `git add -A`)", workingDir)
	}

	response, respErr := provider.Query(
		ctx,
		changes,
		commits,
		ai.WithShortMessageOnly(a.opt.ShortMessageOnly),
		ai.WithEmoji(a.opt.EnableEmoji),
		ai.WithMaxOutputTokens(a.opt.MaxOutputTokens),
	)
	if respErr != nil {
		return respErr
	}

	debug.Printf("prompt:\n%s", response.Prompt)
	debug.Printf("answer:\n%s\n", response.Answer)

	if _, err := fmt.Fprintln(os.Stdout, response.Answer); err != nil {
		return err
	}

	return nil
}
