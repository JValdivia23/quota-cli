package providers

import (
	"context"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

// Provider interface defines the methods all supported AI APIs must implement.
type Provider interface {
	// Name returns the human-readable provider name (e.g. "OpenRouter").
	Name() string

	// Type returns whether the provider is pay-as-you-go or quota-based.
	Type() models.ProviderType

	// Fetch fetches the live usage data using the provided configuration.
	Fetch(ctx context.Context, cfg *models.OpenCodeAuthConfig) (*models.ProviderReport, error)

	// FetchHistory retrieves historical usage data if supported by the provider.
	FetchHistory(ctx context.Context, cfg *models.OpenCodeAuthConfig) ([]models.DailyUsage, error)
}
