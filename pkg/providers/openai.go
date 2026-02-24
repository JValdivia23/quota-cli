package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

// OpenAIProvider implements the OpenAI Quota-based fetcher
type OpenAIProvider struct{}

func (c *OpenAIProvider) Name() string {
	return "OpenAI"
}

func (c *OpenAIProvider) Type() models.ProviderType {
	return models.TypeQuotaBased
}

func (c *OpenAIProvider) Fetch(ctx context.Context, cfg *models.OpenCodeAuthConfig) (*models.ProviderReport, error) {
	token := cfg.GetNestedField("openai", "access")
	if token == "" {
		token = cfg.GetKey(c.Name())
	}

	// If no real token, do not return mock data
	if token == "" || token == "sk-openai-mock" {
		return &models.ProviderReport{
			Name:        c.Name(),
			Type:        models.TypeQuotaBased,
			RefreshTime: "Token Missing or Mock",
		}, fmt.Errorf("invalid or missing OpenAI token")
	}

	accountID := cfg.GetNestedField("openai", "accountId")

	req, err := http.NewRequestWithContext(ctx, "GET", "https://chatgpt.com/backend-api/wham/usage", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if accountID != "" {
		req.Header.Set("ChatGPT-Account-Id", accountID)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return &models.ProviderReport{
			Name:        c.Name(),
			Type:        models.TypeQuotaBased,
			RefreshTime: "Token Expired or " + string(bodyBytes),
		}, fmt.Errorf("API failed with status %d", resp.StatusCode)
	}

	var result struct {
		RateLimit struct {
			PrimaryWindow struct {
				UsedPercent float64 `json:"used_percent"`
				ResetAt     int64   `json:"reset_at"`
			} `json:"primary_window"`
		} `json:"rate_limit"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}

	// Calculate used percentage
	usage := int(result.RateLimit.PrimaryWindow.UsedPercent)
	if usage > 100 {
		usage = 100
	} else if usage < 0 {
		usage = 0
	}
	remaining := 100 - usage

	// Format refresh time
	refreshTimeStr := "Weekly" // fallback
	if result.RateLimit.PrimaryWindow.ResetAt > 0 {
		resetTime := time.Unix(result.RateLimit.PrimaryWindow.ResetAt, 0)
		days := int(time.Until(resetTime).Hours() / 24)
		if days > 0 {
			refreshTimeStr = fmt.Sprintf("Weekly: in %dd (%s)", days, resetTime.Format("01/02"))
		} else {
			refreshTimeStr = fmt.Sprintf("Weekly: %s", resetTime.Format("01/02 15:04"))
		}
	}

	return &models.ProviderReport{
		Name:            c.Name(),
		Type:            models.TypeQuotaBased,
		Remaining:       remaining,
		Entitlement:     100,
		UsagePercentage: usage,
		RefreshTime:     refreshTimeStr,
	}, nil
}

func (c *OpenAIProvider) FetchHistory(ctx context.Context, cfg *models.OpenCodeAuthConfig) ([]models.DailyUsage, error) {
	return nil, nil
}
