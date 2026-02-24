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

// CopilotProvider implements the GitHub Copilot Quota-based fetcher using the OAuth token
// stored in auth.json — no browser cookies required.
type CopilotProvider struct{}

func (c *CopilotProvider) Name() string {
	return "GitHub Copilot"
}

func (c *CopilotProvider) Type() models.ProviderType {
	return models.TypeQuotaBased
}

func (c *CopilotProvider) Fetch(ctx context.Context, cfg *models.OpenCodeAuthConfig) (*models.ProviderReport, error) {
	// Try the nested "github-copilot" key from auth.json
	token := cfg.GetNestedField("github-copilot", "access")
	if token == "" {
		token = cfg.GetKey(c.Name())
	}
	if token == "" {
		return nil, fmt.Errorf("no GitHub Copilot OAuth token found in auth.json")
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
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("copilot API returned status %d", resp.StatusCode)
	}

	var result struct {
		QuotaResetDateUTC string `json:"quota_reset_date_utc"`
		QuotaSnapshots    struct {
			PremiumInteractions struct {
				Entitlement      int     `json:"entitlement"`
				Remaining        int     `json:"remaining"`
				PercentRemaining float64 `json:"percent_remaining"`
				OveragePermitted bool    `json:"overage_permitted"`
			} `json:"premium_interactions"`
		} `json:"quota_snapshots"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse copilot response: %w", err)
	}

	premium := result.QuotaSnapshots.PremiumInteractions
	if premium.Entitlement == 0 {
		return nil, fmt.Errorf("no premium_interactions quota data in response")
	}

	used := premium.Entitlement - premium.Remaining
	usagePct := 0
	if premium.Entitlement > 0 {
		usagePct = (used * 100) / premium.Entitlement
	}

	// Format reset time
	refreshStr := "Monthly"
	if result.QuotaResetDateUTC != "" {
		if resetTime, err := time.Parse(time.RFC3339, result.QuotaResetDateUTC); err == nil {
			daysLeft := int(time.Until(resetTime).Hours() / 24)
			if daysLeft > 0 {
				refreshStr = fmt.Sprintf("Monthly: in %dd (%s)", daysLeft, resetTime.Format("01/02"))
			} else {
				refreshStr = fmt.Sprintf("Monthly: %s", resetTime.Format("01/02"))
			}
		}
	}

	return &models.ProviderReport{
		Name:             c.Name(),
		Type:             c.Type(),
		Remaining:        premium.Remaining,
		Entitlement:      premium.Entitlement,
		UsagePercentage:  usagePct,
		OveragePermitted: premium.OveragePermitted,
		RefreshTime:      refreshStr,
	}, nil
}

func (c *CopilotProvider) FetchHistory(ctx context.Context, cfg *models.OpenCodeAuthConfig) ([]models.DailyUsage, error) {
	return nil, nil
}
