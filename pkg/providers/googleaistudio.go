package providers

import (
	"context"
	"fmt"

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
	return &models.ProviderReport{
		Name:        c.Name(),
		Type:        c.Type(),
		RefreshTime: "(Not Implemented)",
	}, fmt.Errorf("Google AI Studio quota checking is not yet implemented")
}

func (c *GoogleAIStudioProvider) FetchHistory(ctx context.Context, cfg *models.OpenCodeAuthConfig) ([]models.DailyUsage, error) {
	return nil, nil
}
