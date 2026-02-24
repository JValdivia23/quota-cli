package providers

import (
	"context"
	"os"
	"path/filepath"

	"github.com/JValdivia23/quota-cli/pkg/models"
	"golang.org/x/oauth2/google"
)

// GetActiveProviders discovers which providers have credentials available
// and returns only those — no hardcoded assumptions about the user's setup.
func GetActiveProviders(cfg *models.OpenCodeAuthConfig, provFlag string) []Provider {
	// Full catalog of supported providers
	catalog := []Provider{
		&AntigravityProvider{},
		&OpenAIProvider{},
		&OpenRouterProvider{},
		&CopilotProvider{},
		&VertexProvider{},
		&OpenCodeZenProvider{},
	}

	var active []Provider
	for _, p := range catalog {
		// Apply --provider filter if given
		if provFlag != "" && p.Name() != provFlag {
			continue
		}

		if isProviderAvailable(p, cfg) {
			active = append(active, p)
		}
	}

	return active
}

// isProviderAvailable probes each provider's credential sources to determine
// if it should be included in the active set.
func isProviderAvailable(p Provider, cfg *models.OpenCodeAuthConfig) bool {
	switch p.Name() {

	case "Vertex AI":
		_, err := google.FindDefaultCredentials(context.Background(), "https://www.googleapis.com/auth/cloud-platform")
		return err == nil

	case "Antigravity":
		hasAnthropic := cfg.GetNestedField("anthropic", "access") != "" || cfg.GetNestedField("anthropic", "key") != ""
		hasGemini := cfg.GetKey("Gemini CLI") != "" || hasAntigravityConfig(cfg)
		return hasAnthropic || hasGemini

	case "GitHub Copilot":
		return cfg.GetNestedField("github-copilot", "access") != "" ||
			cfg.GetNestedField("github-copilot", "refresh") != ""

	case "OpenCode Zen":
		// Only mark as available if a real token was found in the DB during auth discovery
		t, ok := cfg.RawKeys["opencode-zen-token"].(string)
		return ok && t != ""

	default:
		return cfg.GetKey(p.Name()) != ""
	}
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

// hasOpenCodeZenDBFile checks if the opencode.db file exists (token presence
// is verified by auth.DiscoverOpenCodeAuth at startup, exposed via cfg.RawKeys).
func hasOpenCodeZenDBFile() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	for _, p := range []string{
		filepath.Join(home, ".local", "share", "opencode", "opencode.db"),
		filepath.Join(home, "Library", "Application Support", "opencode", "opencode.db"),
	} {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}
