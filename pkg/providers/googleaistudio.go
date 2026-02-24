package providers

import (
	"context"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

// GoogleAIStudioProvider implements the Google AI Studio fetcher
type GoogleAIStudioProvider struct{}

func (c *GoogleAIStudioProvider) Name() string {
	return "Google AI Studio"
}

func (c *GoogleAIStudioProvider) Type() models.ProviderType {
	return models.TypeQuotaBased
}

func (c *GoogleAIStudioProvider) Fetch(ctx context.Context, cfg *models.OpenCodeAuthConfig) (*models.ProviderReport, error) {
	// Mocking data for the v1 implementation
	return &models.ProviderReport{
		Name:             c.Name(),
		Type:             c.Type(),
		Remaining:        50,
		Entitlement:      100,
		UsagePercentage:  50,
		OveragePermitted: false,
		RefreshTime:      "Tomorrow",
	}, nil
}

func (c *GoogleAIStudioProvider) FetchHistory(ctx context.Context, cfg *models.OpenCodeAuthConfig) ([]models.DailyUsage, error) {
	return nil, nil
}
