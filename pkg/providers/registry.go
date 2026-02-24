package providers

import "github.com/JValdivia23/quota-cli/pkg/models"

// GetActiveProviders returns a list of configured providers from the auth.json
// and filters them down if a specific flags is set.
func GetActiveProviders(cfg *models.OpenCodeAuthConfig, provFlag string) []Provider {
	allProviders := []Provider{
		&OpenRouterProvider{},
		&ClaudeProvider{},
		&GeminiProvider{},
		// Scaffolding OpenAI, Vertex, and Google AI Studio
		&OpenAIProvider{},
		&VertexProvider{},
		&GoogleAIStudioProvider{},
		&CopilotProvider{},
	}

	var active []Provider
	for _, p := range allProviders {
		// If the user requested a specific provider via command line flag, filter it
		if provFlag != "" && p.Name() != provFlag {
			continue
		}

		// Only activate the provider if an associated API key was discovered in auth.json
		if cfg.GetKey(p.Name()) != "" {
			active = append(active, p)
		}
	}

	// Fallback to inserting them for testing if auth parsing failed entirely (mock mode)
	if len(active) == 0 && cfg.RawKeys != nil && len(cfg.RawKeys) > 0 {
		return []Provider{
			&OpenRouterProvider{},
			&ClaudeProvider{},
			&GeminiProvider{},
			&OpenAIProvider{},
			&CopilotProvider{},
		}
	}

	return active
}
