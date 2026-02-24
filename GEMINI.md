# qcli Development Guide

## Project Overview
qcli is a fast, cross-platform tool developed in Go to measure and report on user quotas from various AI providers (Claude, Gemini, Google AI Studio, OpenAI, OpenRouter, Vertex).

## Architecture & Conventions
- **Language**: Go 1.25+
- **CLI Framework**: Cobra (`github.com/spf13/cobra`) and Viper for configuration.
- **Project Structure**:
  - `cmd/`: Contains the CLI commands (e.g., `cmd/quota` for the main executable).
  - `pkg/`: Publicly usable packages (auth, display, models, providers).
  - `internal/`: Internal application logic (e.g., config).
- **Style Guide**:
  - Strictly adhere to standard Go formatting (`gofmt`).
  - Use clear, idiomatic Go naming conventions.
  - Keep commands in the `cmd/` directory minimal, delegating logic to the `pkg/` or `internal/` directories.
  - Implement the `providers` interfaces for all new AI services added.

## Build and Test Commands
- **Run the CLI locally**: `go run main.go [command]`
- **Build the binary**: `go build -o qcli main.go`
- **Format code**: `go fmt ./...`
- **Run tests**: `go test ./...`

## Core Commands
- **`qcli` (Status)**: Default command. Fetches and displays a quick snapshot of all active providers.
- **`qcli report`**: Generates a deep-dive analysis including weighted 7-day usage history and monthly forecasts.
- **`qcli list`**: Lists all providers where local authentication was successfully detected.

## Key Features
- **Prediction Engine**: Located in `pkg/predictor`, uses a weighted-average algorithm with weekend compensation to forecast monthly totals and overage costs.
- **Auth Discovery**: Automatically scans standard OpenCode paths for `auth.json` and `antigravity-accounts.json` to extract OAuth tokens and API keys.
- **Provider Interface**: All providers must implement the `Provider` interface in `pkg/providers`, including the `FetchHistory` method for trend analysis.

## Gemini CLI Instructions
- When adding a new command, place it in `cmd/quota/` and ensure it uses `cobra`.
- When adding a new AI provider, implement it in `pkg/providers/` and register it appropriately in `pkg/providers/registry.go`.
- The `Fetch` method now takes `*models.OpenCodeAuthConfig` as an argument to allow access to nested configuration fields.
- Always verify changes by running `go build -o qcli main.go` and `go test ./...` before considering a task complete.
- Do not check in `.env` files or hardcode API keys. Rely on `pkg/auth` for configuration discovery.
