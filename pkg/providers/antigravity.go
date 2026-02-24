package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

// AntigravityProvider shows Claude and Gemini CLI as a unified "Antigravity" provider,
// matching the opencodebar display with sub-rows and accurate reset times.
type AntigravityProvider struct{}

func (a *AntigravityProvider) Name() string {
	return "Antigravity"
}

func (a *AntigravityProvider) Type() models.ProviderType {
	return models.TypeQuotaBased
}

func (a *AntigravityProvider) Fetch(ctx context.Context, cfg *models.OpenCodeAuthConfig) (*models.ProviderReport, error) {
	var accounts []models.Account
	var idx int

	// --- Claude ---
	claudeToken := cfg.GetNestedField("anthropic", "access")
	if claudeToken == "" {
		claudeToken = cfg.GetKey("Claude")
	}
	if claudeToken != "" {
		acc, err := fetchClaudeAccount(ctx, claudeToken, idx)
		if err == nil {
			accounts = append(accounts, acc)
			idx++
		}
	}

	// --- Gemini CLI ---
	geminiToken := cfg.GetKey("Gemini CLI")
	if geminiToken == "" {
		// Try to refresh from antigravity config
		var refreshErr error
		geminiProvider := &GeminiProvider{}
		geminiToken, refreshErr = geminiProvider.refreshToken(ctx, cfg)
		if refreshErr != nil {
			geminiToken = ""
		}
	}
	if geminiToken != "" {
		acc, err := fetchGeminiAccount(ctx, geminiToken, idx)
		if err == nil {
			accounts = append(accounts, acc)
			idx++
		}
	}

	if len(accounts) == 0 {
		return nil, fmt.Errorf("no Antigravity accounts found (missing anthropic/gemini tokens)")
	}

	return &models.ProviderReport{
		Name:     a.Name(),
		Type:     a.Type(),
		Accounts: accounts,
	}, nil
}

func fetchClaudeAccount(ctx context.Context, token string, idx int) (models.Account, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.anthropic.com/api/oauth/usage", nil)
	if err != nil {
		return models.Account{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return models.Account{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.Account{}, fmt.Errorf("claude API returned status %d", resp.StatusCode)
	}

	var result struct {
		SevenDay struct {
			Utilization float64 `json:"utilization"`
			ResetsAt    string  `json:"resets_at"`
		} `json:"seven_day"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.Account{}, err
	}

	usage := int(result.SevenDay.Utilization)
	remaining := 100 - usage
	if remaining < 0 {
		remaining = 0
	}

	// Format a human-readable reset time
	refreshLabel := "Claude"
	if result.SevenDay.ResetsAt != "" {
		if t, err := time.Parse(time.RFC3339, result.SevenDay.ResetsAt); err == nil {
			hoursLeft := int(time.Until(t).Hours())
			if hoursLeft < 24 {
				refreshLabel = fmt.Sprintf("Claude: in %dh (%s)", hoursLeft, t.Format("15:04"))
			} else {
				refreshLabel = fmt.Sprintf("Claude: in %dd (%s)", hoursLeft/24, t.Format("01/02"))
			}
		}
	}

	return models.Account{
		Index:               idx,
		Email:               refreshLabel,
		Remaining:           remaining,
		Entitlement:         100,
		RemainingPercentage: remaining,
	}, nil
}

func fetchGeminiAccount(ctx context.Context, token string, idx int) (models.Account, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", "https://cloudcode-pa.googleapis.com/v1internal:retrieveUserQuota", nil)
	if err != nil {
		return models.Account{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return models.Account{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.Account{}, fmt.Errorf("gemini API returned status %d", resp.StatusCode)
	}

	var result struct {
		Buckets []struct {
			DisplayName       string  `json:"displayName"`
			RemainingFraction float64 `json:"remainingFraction"`
			ResetTime         string  `json:"resetTime"`
		} `json:"buckets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.Account{}, err
	}
	if len(result.Buckets) == 0 {
		return models.Account{}, fmt.Errorf("no quota buckets found")
	}

	// Use the worst (lowest) remaining fraction
	minFraction := 1.0
	resetTimeStr := ""
	for _, b := range result.Buckets {
		if b.RemainingFraction < minFraction {
			minFraction = b.RemainingFraction
			resetTimeStr = b.ResetTime
		}
	}
	remaining := int(minFraction * 100)
	used := 100 - remaining

	// Format reset label
	refreshLabel := "Gemini"
	if resetTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, resetTimeStr); err == nil {
			hoursLeft := int(time.Until(t).Hours())
			if hoursLeft < 24 {
				refreshLabel = fmt.Sprintf("Gemini: in %dh (%s)", hoursLeft, t.Format("15:04"))
			} else {
				refreshLabel = fmt.Sprintf("Gemini: in %dd (%s)", hoursLeft/24, t.Format("01/02"))
			}
		}
	}

	return models.Account{
		Index:               idx,
		Email:               refreshLabel,
		Remaining:           remaining,
		Entitlement:         100,
		RemainingPercentage: remaining,
		ModelBreakdown:      map[string]int{"used": used},
	}, nil
}

func (a *AntigravityProvider) FetchHistory(ctx context.Context, cfg *models.OpenCodeAuthConfig) ([]models.DailyUsage, error) {
	return nil, nil
}
