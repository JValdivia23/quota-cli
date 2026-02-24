package providers

import (
	"context"
	"os"
	"path/filepath"

	"github.com/JValdivia23/quota-cli/pkg/models"
	"golang.org/x/oauth2/google"
)

// GetActiveProviders returns a list of configured providers from the auth.json
// and filters them down if a specific provider flag is set.
func GetActiveProviders(cfg *models.OpenCodeAuthConfig, provFlag string) []Provider {
	allProviders := []Provider{
		// Multi-account providers (always try — they self-detect sub-tokens)
		&AntigravityProvider{},
		// Single-account providers
		&OpenAIProvider{},
		&OpenRouterProvider{},
		&CopilotProvider{},
		&VertexProvider{},
		&OpenCodeZenProvider{},
	}

	var active []Provider
	for _, p := range allProviders {
		// Filter by --provider flag if provided
		if provFlag != "" && p.Name() != provFlag {
			continue
		}

		switch p.Name() {
		case "Vertex AI":
			// Vertex uses Application Default Credentials dynamically
			_, err := google.FindDefaultCredentials(context.Background(), "https://www.googleapis.com/auth/cloud-platform")
			if err == nil {
				active = append(active, p)
			}

		case "Antigravity":
			// Active if Claude or Gemini token exists
			hasAnthropic := cfg.GetNestedField("anthropic", "access") != "" || cfg.GetKey("Claude") != ""
			hasGemini := cfg.GetKey("Gemini CLI") != "" || hasAntigravityConfig(cfg)
			if hasAnthropic || hasGemini {
				active = append(active, p)
			}

		case "GitHub Copilot":
			// Active if the github-copilot access token exists
			if cfg.GetNestedField("github-copilot", "access") != "" || cfg.GetKey(p.Name()) != "" {
				active = append(active, p)
			}

		case "OpenCode Zen":
			// Active if token found in local SQLite DB or auth.json
			if hasOpenCodeZenToken() || cfg.GetNestedField("zen", "access") != "" {
				active = append(active, p)
			}

		default:
			// All other providers: require a key in auth.json
			if cfg.GetKey(p.Name()) != "" {
				active = append(active, p)
			}
		}
	}

	return active
}

// hasAntigravityConfig checks if an antigravity refresh token is available.
func hasAntigravityConfig(cfg *models.OpenCodeAuthConfig) bool {
	ant, ok := cfg.RawKeys["antigravity"].(map[string]interface{})
	if !ok {
		return false
	}
	_, hasRefresh := ant["refresh_token"]
	return hasRefresh
}

// hasOpenCodeZenToken checks if opencode.db has a token in the control_account table.
func hasOpenCodeZenToken() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	dbPaths := []string{
		filepath.Join(home, ".local", "share", "opencode", "opencode.db"),
		filepath.Join(home, "Library", "Application Support", "opencode", "opencode.db"),
	}
	for _, p := range dbPaths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}
