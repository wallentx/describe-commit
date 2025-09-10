package cli

import (
	"testing"
)

func TestOptions_Validate_CommitHashAndRepo(t *testing.T) {
	tests := []struct {
		name      string
		repoURL   string
		commitHash string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "both empty - should pass",
			repoURL:   "",
			commitHash: "",
			wantErr:   false,
		},
		{
			name:      "only commitHash set - should pass",
			repoURL:   "",
			commitHash: "abc123",
			wantErr:   false,
		},
		{
			name:      "both set - should pass",
			repoURL:   "owner/repo",
			commitHash: "abc123",
			wantErr:   false,
		},
		{
			name:      "only repoURL set - should fail",
			repoURL:   "owner/repo",
			commitHash: "",
			wantErr:   true,
			errMsg:    "commit hash must be provided when using --repo option",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := newOptionsWithDefaults()
			o.RepoURL = tt.repoURL
			o.CommitHash = tt.commitHash
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
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}