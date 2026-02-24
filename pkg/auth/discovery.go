package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

// DiscoverOpenCodeAuth scans all known auth sources and builds a unified config.
// Sources (in priority order):
//  1. OpenCode auth.json (XDG_DATA_HOME, ~/.local/share, ~/Library/Application Support)
//  2. antigravity-accounts.json
//  3. Environment variables (OPENAI_API_KEY, ANTHROPIC_API_KEY, etc.)
func DiscoverOpenCodeAuth() (*models.OpenCodeAuthConfig, error) {
	cfg := &models.OpenCodeAuthConfig{
		RawKeys: make(map[string]interface{}),
	}

	// 1. Load auth.json
	loaded := false
	for _, p := range getAuthJSONPaths() {
		if data, err := os.ReadFile(p); err == nil {
			var raw map[string]interface{}
			if err := json.Unmarshal(data, &raw); err == nil {
				// Merge all keys
				for k, v := range raw {
					cfg.RawKeys[k] = v
				}
				// Also populate typed fields via struct decode
				json.Unmarshal(data, cfg) //nolint:errcheck
				loaded = true
				break
			}
		}
	}

	// 2. Load antigravity-accounts.json (merged under "antigravity")
	for _, p := range getAntigravityPaths() {
		if data, err := os.ReadFile(p); err == nil {
			var antRaw map[string]interface{}
			if err := json.Unmarshal(data, &antRaw); err == nil {
				cfg.RawKeys["antigravity"] = antRaw
				break
			}
		}
	}

	// 3. Pull env variables as a fallback (works even without OpenCode installed)
	injectEnvKeys(cfg)

	// 4. Load opencode DB for OpenCode Zen token
	if token := discoverZenDBToken(); token != "" {
		cfg.RawKeys["opencode-zen-token"] = token
	}

	if !loaded && len(cfg.RawKeys) == 0 {
		return nil, fmt.Errorf("no provider credentials found (no auth.json and no environment variables)")
	}

	return cfg, nil
}

// injectEnvKeys reads standard environment variables and injects them into RawKeys
// so they work exactly like auth.json entries without any OpenCode dependency.
func injectEnvKeys(cfg *models.OpenCodeAuthConfig) {
	envMap := map[string]string{
		"OPENAI_API_KEY":     "openai.key",
		"ANTHROPIC_API_KEY":  "anthropic.key",
		"OPENROUTER_API_KEY": "openrouter.key",
		"GEMINI_API_KEY":     "gemini.key",
		"GOOGLE_API_KEY":     "google.key",
		"GITHUB_TOKEN":       "github-copilot.access",
		"COPILOT_TOKEN":      "github-copilot.access",
	}

	for envVar, keyPath := range envMap {
		if val := os.Getenv(envVar); val != "" {
			// Only inject if not already set by auth.json
			if _, exists := cfg.RawKeys[keyPath]; !exists {
				cfg.RawKeys[keyPath] = val
			}
		}
	}

	// Also check OPENAI_API_KEY → nested openai.access (the format auth.json uses)
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		if _, ok := cfg.RawKeys["openai"]; !ok {
			cfg.RawKeys["openai"] = map[string]interface{}{"access": key}
		}
	}
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		if _, ok := cfg.RawKeys["anthropic"]; !ok {
			cfg.RawKeys["anthropic"] = map[string]interface{}{"access": key}
		}
	}
}

// discoverZenDBToken looks for an active control_account token in the opencode SQLite DB.
func discoverZenDBToken() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
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
		var token string
		// Try active account first, then any account
		for _, query := range []string{
			`SELECT access_token FROM control_account WHERE active=1 ORDER BY time_updated DESC LIMIT 1`,
			`SELECT access_token FROM control_account ORDER BY time_updated DESC LIMIT 1`,
		} {
			if err := db.QueryRow(query).Scan(&token); err == nil && token != "" {
				db.Close()
				return token
			}
		}
		db.Close()
	}
	return ""
}

func getAuthJSONPaths() []string {
	var paths []string
	if home, err := os.UserHomeDir(); err == nil {
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			paths = append(paths, filepath.Join(xdgData, "opencode", "auth.json"))
		}
		paths = append(paths,
			filepath.Join(home, ".local", "share", "opencode", "auth.json"),
			filepath.Join(home, "Library", "Application Support", "opencode", "auth.json"),
		)
	}
	return paths
}

func getAntigravityPaths() []string {
	var paths []string
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths,
			filepath.Join(home, ".config", "opencode", "antigravity-accounts.json"),
			filepath.Join(home, ".local", "share", "opencode", "antigravity-accounts.json"),
		)
	}
	return paths
}
