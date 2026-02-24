package providers

import (
	"context"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

// VertexProvider implements the Vertex AI Quota-based fetcher
type VertexProvider struct{}

func (c *VertexProvider) Name() string {
	return "Vertex AI"
}

func (c *VertexProvider) Type() models.ProviderType {
	return models.TypeQuotaBased
}

func (c *VertexProvider) Fetch(ctx context.Context, cfg *models.OpenCodeAuthConfig) (*models.ProviderReport, error) {
	// Mocking data for the v1 implementation
	return &models.ProviderReport{
		Name:             c.Name(),
		Type:             c.Type(),
		Remaining:        99,
		Entitlement:      100,
		UsagePercentage:  1,
		OveragePermitted: false,
		RefreshTime:      "in 1d",
	}, nil
}

func (c *VertexProvider) FetchHistory(ctx context.Context, cfg *models.OpenCodeAuthConfig) ([]models.DailyUsage, error) {
	return nil, nil
}
