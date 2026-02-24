package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

// ClaudeProvider implements the Claude Quota-based fetcher
type ClaudeProvider struct{}

func (c *ClaudeProvider) Name() string {
	return "Claude"
}

func (c *ClaudeProvider) Type() models.ProviderType {
	return models.TypeQuotaBased
}

func (c *ClaudeProvider) Fetch(ctx context.Context, cfg *models.OpenCodeAuthConfig) (*models.ProviderReport, error) {
	apiKey := cfg.GetKey(c.Name())
	if apiKey == "" {
		return nil, fmt.Errorf("no API key provided")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.anthropic.com/api/oauth/usage", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("claude API returned status %d", resp.StatusCode)
	}

	var result struct {
		SevenDay struct {
			Utilization float64 `json:"utilization"`
			ResetsAt    string  `json:"resets_at"`
		} `json:"seven_day"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	usage := int(result.SevenDay.Utilization)
	remaining := 100 - usage
	if remaining < 0 {
		remaining = 0
	}

	return &models.ProviderReport{
		Name:            c.Name(),
		Type:            c.Type(),
		Remaining:       remaining,
		Entitlement:     100,
		UsagePercentage: usage,
		RefreshTime:     result.SevenDay.ResetsAt,
	}, nil
}

func (c *ClaudeProvider) FetchHistory(ctx context.Context, cfg *models.OpenCodeAuthConfig) ([]models.DailyUsage, error) {
	return nil, nil
}
