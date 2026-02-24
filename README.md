# quota-cli (qcli)

A fast, cross-platform command-line tool written in Go to measure, report on, and forecast your API usage quotas across various AI providers, including Claude, Gemini, Google AI Studio, OpenAI, OpenRouter, and Vertex AI.

## Features

- **Multi-Provider Support**: Seamlessly displays your usage and quota information for leading AI models in a single unified view.
- **Reporting & Forecasting**: Uses a weighted-average algorithm (with weekend compensation) to predict monthly totals, overage costs, and multi-day trends.
- **Auto-Discovery for Auth**: Automatically detects standards OpenCode token paths (`auth.json`, `antigravity-accounts.json`) to safely discover API keys—no manual `.env` file configuration needed.

## Installation

### Prerequisites

Ensure you have [Go 1.25+](https://go.dev/) or newer installed.

### Method 1: Pre-compiled Binaries (Easiest for most users)

You can download a pre-compiled, ready-to-run binary for Windows (`.exe`), macOS, and Linux directly from our [GitHub Releases page](https://github.com/JValdivia23/quota-cli/releases). 
Just download the right file for your operating system, extract it, and run `qcli`!

### Method 2: Homebrew (macOS & Linux)

If you are using [Homebrew](https://brew.sh/), you can easily install the CLI using:
```bash
brew tap JValdivia23/tap
brew install quota-cli
```
*(Note: To enable Homebrew support, a separate `homebrew-tap` repository needs to be created first).*

### Method 3: Via Go Install

If you already have Go installed, you can simply run:

```bash
go install github.com/JValdivia23/quota-cli@latest
```
*Note: Make sure your `GOPATH/bin` is in your system's `$PATH` so you can use the `quota-cli` command globally.*

### Method 4: Build from Source

1. Clone the repository
   ```bash
   git clone https://github.com/JValdivia23/quota-cli.git
   cd quota-cli
   ```

2. Build the executable
   ```bash
   go build -o qcli main.go
   ```

3. (Optional) Move the compiled binary (`qcli`) into a directory inside your system's PATH. For example:
   ```bash
   mv qcli /usr/local/bin/
   ```

## Usage

Simply run the commands via the output binary:

- **`qcli`**: The default command. Fetches and displays a live snapshot of all your active AI providers.
- **`qcli report`**: Generates a deep-dive analysis, showing weighted 7-day usage history and monthly forecasts.
- **`qcli list`**: Lists all AI providers where local authentication is successfully detected.

## Project Structure

- `cmd/`: Core Cobra CLI commands (`cmd/quota` module).
- `pkg/`: Shared logic to consume data from API providers, models, authentication flow, and predictive engine.
- `internal/`: Internal application configurations and hidden behavior.

## Development

- **Run locally**: `go run main.go [command]`
- **Build**: `go build -o qcli main.go`
- **Format code**: `go fmt ./...`
- **Run tests**: `go test ./...`

## License

*(License to be determined - e.g. MIT / Apache 2.0)*
