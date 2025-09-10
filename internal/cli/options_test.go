package cli

import (
	"testing"
)

func TestOptions_Validate_CommitHashAndRepo(t *testing.T) {
	tests := []struct {
		name       string
		repoURL    string
		commitHash string
		branch     string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "all empty - should pass",
			repoURL:    "",
			commitHash: "",
			branch:     "",
			wantErr:    false,
		},
		{
			name:       "only commitHash set - should pass",
			repoURL:    "",
			commitHash: "abc123",
			branch:     "",
			wantErr:    false,
		},
		{
			name:       "commitHash and branch set (no repo) - should pass",
			repoURL:    "",
			commitHash: "abc123",
			branch:     "main",
			wantErr:    false,
		},
		{
			name:       "all set - should pass",
			repoURL:    "owner/repo",
			commitHash: "abc123",
			branch:     "develop",
			wantErr:    false,
		},
		{
			name:       "repo and branch set (no commitHash) - should fail",
			repoURL:    "owner/repo",
			commitHash: "",
			branch:     "main",
			wantErr:    true,
			errMsg:     "commit hash must be provided when using --repo option",
		},
		{
			name:       "only repoURL set - should fail",
			repoURL:    "owner/repo",
			commitHash: "",
			branch:     "",
			wantErr:    true,
			errMsg:     "commit hash must be provided when using --repo option",
		},
		{
			name:       "only branch set - should pass",
			repoURL:    "",
			commitHash: "",
			branch:     "feature-branch",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := newOptionsWithDefaults()
			o.RepoURL = tt.repoURL
			o.CommitHash = tt.commitHash
			o.Branch = tt.branch
			// Set a valid AI provider to avoid other validation errors
			o.AIProviderName = "gemini"
			o.Providers.Gemini.ApiKey = "test-key"
			o.Providers.Gemini.ModelName = "test-model"

			err := o.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				if err.Error() != tt.errMsg {
					t.Errorf("expected error message %q, got %q", tt.errMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}
