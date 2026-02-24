package models

// ProviderType categorizes how the provider bills its usage.
type ProviderType string

const (
	TypePayAsYouGo  = "pay-as-you-go"
	TypeQuotaBased  = "quota-based"
	TypeTokensBased = "tokens-based"
)

// ProviderReport contains all unified metrics for a single provider.
type ProviderReport struct {
	Name string       `json:"name"`
	Type ProviderType `json:"type"`

	// Quota-based metrics
	Remaining        int    `json:"remaining,omitempty"`
	Entitlement      int    `json:"entitlement,omitempty"`
	UsagePercentage  int    `json:"usagePercentage,omitempty"`
	OveragePermitted bool   `json:"overagePermitted,omitempty"`
	RefreshTime      string `json:"refreshTime,omitempty"`

	// Error state (non-fatal: provider was found but fetch failed)
	ErrorMsg string `json:"error,omitempty"`

	// Pay-As-You-Go metrics
	Cost float64 `json:"cost,omitempty"`

	// Tokens-based metrics
	TokensUsed int64 `json:"tokensUsed,omitempty"`

	// Multiple accounts (e.g. Gemini CLI)
	Accounts []Account `json:"accounts,omitempty"`

	// History and Prediction (New)
	History    []DailyUsage      `json:"history,omitempty"`
	Prediction *PredictionReport `json:"prediction,omitempty"`
}

// DailyUsage represents usage for a specific day.
type DailyUsage struct {
	Date             string  `json:"date"`
	IncludedRequests float64 `json:"includedRequests"`
	BilledAmount     float64 `json:"billedAmount"`
}

// PredictionReport holds the forecasted usage metrics.
type PredictionReport struct {
	PredictedMonthlyRequests float64 `json:"predictedMonthlyRequests"`
	PredictedExtraCost       float64 `json:"predictedExtraCost"`
	Confidence               string  `json:"confidence"` // Low, Medium, High
}

// Account holds metadata for providers with multiple local credentials.
type Account struct {
	Index               int            `json:"index"`
	Email               string         `json:"email"`
	AccountID           string         `json:"accountId"`
	Remaining           int            `json:"remaining"`
	Entitlement         int            `json:"entitlement"`
	RemainingPercentage int            `json:"remainingPercentage"`
	ModelBreakdown      map[string]int `json:"modelBreakdown"`
}

// OpenCodeAuthConfig models the structure of auth.json used by OpenCode.
type OpenCodeAuthConfig struct {
	OpenRouterKey string `json:"openrouter.key,omitempty"`
	ClaudeKey     string `json:"claude.key,omitempty"`
	GeminiKey     string `json:"gemini.key,omitempty"`
	CopilotToken  string `json:"copilot.token,omitempty"`
	// add other keys dynamically or explicitly as required.
	RawKeys map[string]interface{} `json:"-"`
}

// GetKey attempts to find the corresponding API key for a provider name.
func (cfg *OpenCodeAuthConfig) GetKey(providerName string) string {
	if cfg.RawKeys == nil {
		return ""
	}

	// Try old flat format first
	var flatKey string
	switch providerName {
	case "OpenRouter":
		flatKey = "openrouter.key"
	case "Claude":
		flatKey = "claude.key"
	case "Gemini CLI":
		flatKey = "gemini.key"
	case "OpenAI":
		flatKey = "openai.key"
	case "Vertex AI":
		flatKey = "vertex.key"
	case "Google AI Studio":
		flatKey = "googleaistudio.key"
	case "GitHub Copilot":
		flatKey = "copilot.token"
	}

	if flatKey != "" {
		if val := extractString(cfg.RawKeys, flatKey); val != "" {
			return val
		}
	}

	// Try nested structure (e.g., {"openai": {"key": "..."}})
	var nestedKey string
	switch providerName {
	case "OpenRouter":
		nestedKey = "openrouter"
	case "Claude":
		nestedKey = "claude"
	case "Gemini CLI":
		nestedKey = "gemini"
	case "OpenAI":
		nestedKey = "openai"
	case "Vertex AI":
		nestedKey = "vertex"
	case "Google AI Studio":
		nestedKey = "google" // Users often use "google" or "google-custom"
	case "GitHub Copilot":
		nestedKey = "github-copilot"
	}

	if nestedKey != "" {
		if val, ok := cfg.RawKeys[nestedKey].(map[string]interface{}); ok {
			// Check common fields in nested structure
			for _, k := range []string{"key", "access", "refresh", "token"} {
				if s, ok := val[k].(string); ok && s != "" {
					return s
				}
			}
		}
	}

	// Extra fallback for Google custom keys
	if providerName == "Google AI Studio" {
		if val, ok := cfg.RawKeys["google-custom"].(map[string]interface{}); ok {
			if s, ok := val["key"].(string); ok {
				return s
			}
		}
	}

	return ""
}

// GetNestedField retrieves a specific field from a nested provider configuration.
func (cfg *OpenCodeAuthConfig) GetNestedField(providerKey, field string) string {
	if cfg.RawKeys == nil {
		return ""
	}
	if val, ok := cfg.RawKeys[providerKey].(map[string]interface{}); ok {
		if s, ok := val[field].(string); ok {
			return s
		}
	}
	return ""
}

func extractString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if s, isStr := val.(string); isStr {
			return s
		}
	}
	return ""
}
