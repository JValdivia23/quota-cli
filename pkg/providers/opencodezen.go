package providers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

// OpenCodeZenProvider fetches pay-as-you-go cost from the OpenCode Zen (api.z.ai) service.
// Token is read from the opencode SQLite database (control_account table) or auth.json.
type OpenCodeZenProvider struct{}

func (o *OpenCodeZenProvider) Name() string {
	return "OpenCode Zen"
}

func (o *OpenCodeZenProvider) Type() models.ProviderType {
	return models.TypePayAsYouGo
}

func (o *OpenCodeZenProvider) Fetch(ctx context.Context, cfg *models.OpenCodeAuthConfig) (*models.ProviderReport, error) {
	token, err := o.findToken(cfg)
	if err != nil || token == "" {
		return nil, fmt.Errorf("no OpenCode Zen token found: %w", err)
	}

	// Fetch current month total cost from model-usage endpoint
	cost, err := o.fetchModelUsageCost(ctx, token)
	if err != nil {
		return nil, err
	}

	return &models.ProviderReport{
		Name: o.Name(),
		Type: o.Type(),
		Cost: cost,
	}, nil
}

func (o *OpenCodeZenProvider) findToken(cfg *models.OpenCodeAuthConfig) (string, error) {
	// 1. Try local SQLite DB (opencode.db control_account table)
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dbPaths := []string{
		filepath.Join(home, ".local", "share", "opencode", "opencode.db"),
		filepath.Join(home, "Library", "Application Support", "opencode", "opencode.db"),
	}
	for _, dbPath := range dbPaths {
		if _, err := os.Stat(dbPath); err != nil {
			continue
		}
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			continue
		}
		defer db.Close()

		var token string
		err = db.QueryRow(`SELECT access_token FROM control_account WHERE active=1 ORDER BY time_updated DESC LIMIT 1`).Scan(&token)
		if err == nil && token != "" {
			return token, nil
		}
		// Try any account
		err = db.QueryRow(`SELECT access_token FROM control_account ORDER BY time_updated DESC LIMIT 1`).Scan(&token)
		if err == nil && token != "" {
			return token, nil
		}
	}

	// 2. Fallback: check auth.json for a "zen" or "z.ai" key
	token := cfg.GetNestedField("zen", "access")
	if token == "" {
		token = cfg.GetNestedField("opencode", "access")
	}
	if token != "" {
		return token, nil
	}

	return "", fmt.Errorf("token not found in DB or auth.json")
}

func (o *OpenCodeZenProvider) fetchModelUsageCost(ctx context.Context, token string) (float64, error) {
	// Get current month window
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	url := fmt.Sprintf("https://api.z.ai/api/monitor/usage/model-usage?start=%s&end=%s",
		startOfMonth.Format("2006-01-02"),
		now.Format("2006-01-02"),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("zen API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse total cost from model-usage response
	var result struct {
		Data []struct {
			Cost float64 `json:"cost"`
		} `json:"data"`
		Total struct {
			Cost float64 `json:"cost"`
		} `json:"total"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		// Try flat total
		var flat struct {
			Cost float64 `json:"cost"`
		}
		if err2 := json.Unmarshal(body, &flat); err2 == nil && flat.Cost > 0 {
			return flat.Cost, nil
		}
		return 0, fmt.Errorf("failed to parse zen response: %s", string(body))
	}

	if result.Total.Cost > 0 {
		return result.Total.Cost, nil
	}
	var totalCost float64
	for _, d := range result.Data {
		totalCost += d.Cost
	}
	return totalCost, nil
}

func (o *OpenCodeZenProvider) FetchHistory(ctx context.Context, cfg *models.OpenCodeAuthConfig) ([]models.DailyUsage, error) {
	return nil, nil
}
