package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

type openrouterResponse struct {
	Data struct {
		TotalUsage float64 `json:"total_usage"`
	} `json:"data"`
}

// OpenRouterProvider implements the OpenRouter Pay-As-You-Go API fetcher
type OpenRouterProvider struct{}

func (o *OpenRouterProvider) Name() string {
	return "OpenRouter"
}

func (o *OpenRouterProvider) Type() models.ProviderType {
	return models.TypePayAsYouGo
}

func (o *OpenRouterProvider) Fetch(ctx context.Context, cfg *models.OpenCodeAuthConfig) (*models.ProviderReport, error) {
	apiKey := cfg.GetKey(o.Name())
	if apiKey == "" {
		return nil, fmt.Errorf("no API key provided")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://openrouter.ai/api/v1/credits", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openrouter API returned status %d", resp.StatusCode)
	}

	var orResp openrouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&orResp); err != nil {
		return nil, err
	}

	return &models.ProviderReport{
		Name: o.Name(),
		Type: o.Type(),
		Cost: orResp.Data.TotalUsage,
	}, nil
}

func (o *OpenRouterProvider) FetchHistory(ctx context.Context, cfg *models.OpenCodeAuthConfig) ([]models.DailyUsage, error) {
	return nil, nil
}
