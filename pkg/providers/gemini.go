package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

// GeminiProvider implements the Gemini CLI API fetcher
type GeminiProvider struct{}

func (g *GeminiProvider) Name() string {
	return "Gemini CLI"
}

func (g *GeminiProvider) Type() models.ProviderType {
	return models.TypeQuotaBased
}

func (g *GeminiProvider) Fetch(ctx context.Context, cfg *models.OpenCodeAuthConfig) (*models.ProviderReport, error) {
	// 1. Try to get access token from auth.json
	accessToken := cfg.GetKey(g.Name())

	// 2. If missing, try to refresh using antigravity config
	if accessToken == "" {
		var err error
		accessToken, err = g.refreshToken(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to get/refresh Gemini token: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://cloudcode-pa.googleapis.com/v1internal:retrieveUserQuota", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// Try refresh once
		accessToken, err = g.refreshToken(ctx, cfg)
		if err == nil {
			// Retry request
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err = client.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API returned status %d", resp.StatusCode)
	}

	var result struct {
		Buckets []struct {
			DisplayName       string  `json:"displayName"`
			RemainingFraction float64 `json:"remainingFraction"`
		} `json:"buckets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Buckets) == 0 {
		return nil, fmt.Errorf("no quota buckets found")
	}

	// Use the lowest remaining fraction as the primary metric
	minFraction := 1.0
	for _, b := range result.Buckets {
		if b.RemainingFraction < minFraction {
			minFraction = b.RemainingFraction
		}
	}

	usage := int((1.0 - minFraction) * 100)
	remaining := int(minFraction * 100)

	return &models.ProviderReport{
		Name:            g.Name(),
		Type:            g.Type(),
		Remaining:       remaining,
		Entitlement:     100,
		UsagePercentage: usage,
	}, nil
}

func (g *GeminiProvider) FetchHistory(ctx context.Context, cfg *models.OpenCodeAuthConfig) ([]models.DailyUsage, error) {
	return nil, nil
}

func (g *GeminiProvider) refreshToken(ctx context.Context, cfg *models.OpenCodeAuthConfig) (string, error) {
	ant, ok := cfg.RawKeys["antigravity"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("antigravity config not found")
	}

	clientID := extractString(ant, "client_id")
	clientSecret := extractString(ant, "client_secret")
	refreshToken := extractString(ant, "refresh_token")

	if clientID == "" || refreshToken == "" {
		return "", fmt.Errorf("missing client_id or refresh_token in antigravity config")
	}

	// In real implementation, clientSecret might be optional or hardcoded if it's a public client
	// For Gemini CLI, it's often hardcoded in the app.

	data := fmt.Sprintf("grant_type=refresh_token&client_id=%s&client_secret=%s&refresh_token=%s",
		clientID, clientSecret, refreshToken)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token refresh failed with status %d", resp.StatusCode)
	}

	var res struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.AccessToken, nil
}

func extractString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}
