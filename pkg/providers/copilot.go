package providers

import (
	"context"
	"fmt"
	"net/http"

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

	// Copilot quota requires cookie-based parsing or explicit endpoint handling that isn't fully supported yet
	return &models.ProviderReport{
		Name:        c.Name(),
		Type:        c.Type(),
		RefreshTime: "(Browser Auth Required)",
	}, fmt.Errorf("copilot direct OAuth quota not fully implemented")
}

func (c *CopilotProvider) FetchHistory(ctx context.Context, cfg *models.OpenCodeAuthConfig) ([]models.DailyUsage, error) {
	return nil, nil
}
