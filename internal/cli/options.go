package cli

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gh.tarampamp.am/describe-commit/internal/ai"
	"gh.tarampamp.am/describe-commit/internal/config"
)

// options represents the command-line options. this struct should be used ONLY in this package (do not try to pass
// it somewhere else).
type options struct {
	ShortMessageOnly    bool
	CommitHistoryLength int64
	EnableEmoji         bool
	MaxOutputTokens     int64
	MaxRetries          uint
	RetryDelay          time.Duration
	AIProviderName      string
	CommitHash          string
	RepoURL             string
	Branch              string

	Providers struct {
		Gemini     struct{ ApiKey, ModelName, BaseURL string }
		OpenAI     struct{ ApiKey, ModelName, BaseURL string }
		OpenRouter struct{ ApiKey, ModelName, BaseURL string }
		Anthropic  struct{ ApiKey, ModelName, BaseURL string }
	}
}

func newOptionsWithDefaults() options {
	var opt = options{
		CommitHistoryLength: 20,  //nolint:mnd
		MaxOutputTokens:     500, //nolint:mnd
		MaxRetries:          5,   //nolint:mnd
		RetryDelay:          time.Second,
		AIProviderName:      ai.ProviderGemini, // due to its free
	}

	// https://ai.google.dev/gemini-api/docs/models
	opt.Providers.Gemini.ModelName = "gemini-2.5-flash"
	// https://developers.openai.com/api/docs/models
	opt.Providers.OpenAI.ModelName = "gpt-4.1-nano"
	// https://openrouter.ai/api/v1/models
	opt.Providers.OpenRouter.ModelName = "google/gemma-4-31b-it:free"
	// https://platform.claude.com/docs/en/about-claude/models/overview
	opt.Providers.Anthropic.ModelName = "claude-haiku-4-5-20251001"

	return opt
}

// UpdateFromConfigFile loads the configuration from the file(s) and applies it to the options.
// The values loaded from the earlier files will be overridden by those from the later files, with the last
// file taking the highest priority.
// Missing files and directories are ignored.
func (o *options) UpdateFromConfigFile(filePath []string) error {
	if len(filePath) == 0 {
		return nil
	}

	var cfg config.Config

	for _, path := range filePath {
		if path == "" {
			continue // skip empty paths
		}

		if stat, err := os.Stat(path); err != nil || stat.IsDir() {
			continue // skip missing files and directories
		}

		if err := cfg.FromFile(path); err != nil {
			return fmt.Errorf("failed to load the configuration file: %w", err)
		}
	}

	return o.applyConfig(cfg)
}

func (o *options) applyConfig(cfg config.Config) error {
	setIfSourceNotNil(&o.ShortMessageOnly, cfg.ShortMessageOnly)
	setIfSourceNotNil(&o.CommitHistoryLength, cfg.CommitHistoryLength)
	setIfSourceNotNil(&o.EnableEmoji, cfg.EnableEmoji)
	setIfSourceNotNil(&o.MaxOutputTokens, cfg.MaxOutputTokens)
	setIfSourceNotNil(&o.MaxRetries, cfg.MaxRetries)
	setIfSourceNotNil(&o.AIProviderName, cfg.AIProviderName)
	setIfSourceNotNil(&o.CommitHash, cfg.CommitHash)
	setIfSourceNotNil(&o.RepoURL, cfg.RepoURL)
	setIfSourceNotNil(&o.Branch, cfg.Branch)

	if err := o.applyRetryDelay(cfg.RetryDelay); err != nil {
		return err
	}

	if sub := cfg.Gemini; sub != nil {
		setProviderConfig(&o.Providers.Gemini.ApiKey, &o.Providers.Gemini.ModelName, &o.Providers.Gemini.BaseURL,
			sub.ApiKey, sub.ModelName, sub.BaseURL)
	}

	if sub := cfg.OpenAI; sub != nil {
		setProviderConfig(&o.Providers.OpenAI.ApiKey, &o.Providers.OpenAI.ModelName, &o.Providers.OpenAI.BaseURL,
			sub.ApiKey, sub.ModelName, sub.BaseURL)
	}

	if sub := cfg.OpenRouter; sub != nil {
		setProviderConfig(
			&o.Providers.OpenRouter.ApiKey, &o.Providers.OpenRouter.ModelName, &o.Providers.OpenRouter.BaseURL,
			sub.ApiKey, sub.ModelName, sub.BaseURL,
		)
	}

	if sub := cfg.Anthropic; sub != nil {
		setProviderConfig(
			&o.Providers.Anthropic.ApiKey, &o.Providers.Anthropic.ModelName, &o.Providers.Anthropic.BaseURL,
			sub.ApiKey, sub.ModelName, sub.BaseURL,
		)
	}

	return nil
}

func (o *options) applyRetryDelay(delay *string) error {
	if delay == nil || *delay == "" {
		return nil
	}

	dur, parseErr := time.ParseDuration(*delay)
	if parseErr != nil {
		return fmt.Errorf("invalid retryDelay value %q: %w", *delay, parseErr)
	}

	o.RetryDelay = dur

	return nil
}

func setProviderConfig(
	apiKeyTarget, modelNameTarget, baseURLTarget *string,
	apiKeySource, modelNameSource, baseURLSource *string,
) {
	setIfSourceNotNil(apiKeyTarget, apiKeySource)
	setIfSourceNotNil(modelNameTarget, modelNameSource)
	setIfSourceNotNil(baseURLTarget, baseURLSource)
}

// setIfSourceNotNil sets the target value to the source value if both are not nil.
func setIfSourceNotNil[T any](target, source *T) {
	if target == nil || source == nil {
		return
	}

	*target = *source
}

func (o *options) Validate() error {
	if o.MaxOutputTokens <= 1 {
		return errors.New("max output tokens must be greater than 1")
	}

	if v := o.AIProviderName; !ai.IsProviderSupported(v) {
		return fmt.Errorf("unsupported AI provider: %s", v)
	}

	// If repo URL is provided, commit hash must also be provided
	if o.RepoURL != "" && o.CommitHash == "" {
		return errors.New("commit hash must be provided when using --repo option")
	}

	return o.validateProviderCredentials()
}

// validateProviderCredentials validates the API keys and model names for the selected provider.
func (o *options) validateProviderCredentials() error {
	switch o.AIProviderName {
	case ai.ProviderGemini:
		return o.validateGemini()
	case ai.ProviderOpenAI:
		return o.validateOpenAI()
	case ai.ProviderOpenRouter:
		return o.validateOpenRouter()
	case ai.ProviderAnthropic:
		return o.validateAnthropic()
	default:
		return fmt.Errorf("unsupported AI provider: %s", o.AIProviderName)
	}
}

func (o *options) validateGemini() error {
	if o.Providers.Gemini.ApiKey == "" {
		return errors.New("gemini API key is required")
	}

	if o.Providers.Gemini.ModelName == "" {
		return errors.New("gemini model name is required")
	}

	return nil
}

func (o *options) validateOpenAI() error {
	if o.Providers.OpenAI.ApiKey == "" {
		return errors.New("missing OpenAI API key")
	}

	if o.Providers.OpenAI.ModelName == "" {
		return errors.New("missing OpenAI model name")
	}

	return nil
}

func (o *options) validateOpenRouter() error {
	if o.Providers.OpenRouter.ApiKey == "" {
		return errors.New("missing OpenRouter API key")
	}

	if o.Providers.OpenRouter.ModelName == "" {
		return errors.New("missing OpenRouter model name")
	}

	return nil
}

func (o *options) validateAnthropic() error {
	if o.Providers.Anthropic.ApiKey == "" {
		return errors.New("missing Anthropic API key")
	}

	if o.Providers.Anthropic.ModelName == "" {
		return errors.New("missing Anthropic model name")
	}

	return nil
}
