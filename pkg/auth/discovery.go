package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

// DiscoverOpenCodeAuth locating OpenCode's auth.json across multiple paths.
func DiscoverOpenCodeAuth() (*models.OpenCodeAuthConfig, error) {
	paths := getCandidatePaths()

	var validPath string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			validPath = p
			break
		}
	}

	if validPath == "" {
		return nil, fmt.Errorf("auth.json not found in any standard OpenCode paths")
	}

	data, err := os.ReadFile(validPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", validPath, err)
	}

	var cfg models.OpenCodeAuthConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", validPath, err)
	}

	// Also parse into a raw map to capture dynamic keys
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	cfg.RawKeys = raw

	// Also look for antigravity-accounts.json
	antigravityPaths := getAntigravityPaths()
	for _, p := range antigravityPaths {
		if data, err := os.ReadFile(p); err == nil {
			var antRaw map[string]interface{}
			if err := json.Unmarshal(data, &antRaw); err == nil {
				// Merge antigravity keys into RawKeys under "antigravity" key or directly
				cfg.RawKeys["antigravity"] = antRaw
			}
			break
		}
	}

	return &cfg, nil
}

func getCandidatePaths() []string {
	var paths []string

	if home, err := os.UserHomeDir(); err == nil {
		// XDG_DATA_HOME fallback
		xdgData := os.Getenv("XDG_DATA_HOME")
		if xdgData != "" {
			paths = append(paths, filepath.Join(xdgData, "opencode", "auth.json"))
		}

		// .local/share fallback
		paths = append(paths, filepath.Join(home, ".local", "share", "opencode", "auth.json"))

		// macOS native fallback
		paths = append(paths, filepath.Join(home, "Library", "Application Support", "opencode", "auth.json"))
	}

	return paths
}

func getAntigravityPaths() []string {
	var paths []string

	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "opencode", "antigravity-accounts.json"))
		paths = append(paths, filepath.Join(home, ".local", "share", "opencode", "antigravity-accounts.json"))
	}

	return paths
}
