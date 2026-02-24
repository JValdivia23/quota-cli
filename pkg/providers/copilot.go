package providers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

// CopilotProvider implements the GitHub Copilot Quota-based fetcher
type CopilotProvider struct{}

func (c *CopilotProvider) Name() string {
	return "GitHub Copilot"
}

func (c *CopilotProvider) Type() models.ProviderType {
	return models.TypeQuotaBased
}

func (c *CopilotProvider) Fetch(ctx context.Context, cfg *models.OpenCodeAuthConfig) (*models.ProviderReport, error) {
	token := cfg.GetKey(c.Name())
	if token == "" {
		return nil, fmt.Errorf("no Copilot token provided")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/copilot_internal/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("copilot API returned status %d", resp.StatusCode)
	}

	// The /copilot_internal/user endpoint returns info about the user.
	// In some versions, it might include quota info.
	// For now, let's assume it returns a successful response and we use a placeholder or 
	// try to parse what we can.
	
	// Mocking successful fetch for now as individual copilot quota is often not exposed via this API directly 
	// without additional headers or from different endpoints.
	return &models.ProviderReport{
		Name:             c.Name(),
		Type:             c.Type(),
		Remaining:        113,
		Entitlement:      300,
		UsagePercentage:  62,
		OveragePermitted: true,
		RefreshTime:      "Monthly",
	}, nil
}

// FetchHistory provides historical usage data for prediction.
func (c *CopilotProvider) FetchHistory(ctx context.Context, cfg *models.OpenCodeAuthConfig) ([]models.DailyUsage, error) {
	// Mock history for testing prediction logic: last 7 days.
	// In a real implementation, we would call /settings/billing/copilot_usage_table.
	today := time.Now().UTC()
	history := []models.DailyUsage{}
	
	// Mock daily usage (around 10-15 requests per day)
	mockUsages := []float64{12, 15, 8, 2, 1, 14, 11} // Recent to oldest
	
	for i, usage := range mockUsages {
		dateStr := today.AddDate(0, 0, -(i + 1)).Format("2006-01-02")
		history = append(history, models.DailyUsage{
			Date:             dateStr,
			IncludedRequests: usage,
		})
	}
	
	return history, nil
}
