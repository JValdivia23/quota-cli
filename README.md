# qcli — AI Quota CLI

A fast, cross-platform CLI written in Go that **auto-discovers your AI provider credentials** and displays real-time quota and cost information in a single unified table.

Supports: **Claude · ChatGPT/OpenAI · GitHub Copilot · Gemini CLI · OpenRouter · Vertex AI · OpenCode Zen**

```
Provider           Refresh                  Use    Key Metrics
────────────────   ──────────────────────   ────   ────────────────────
GitHub Copilot     Monthly: in 4d (03/01)   62%    113/300 remaining
OpenAI             Weekly: in 6d (03/03)    0%     100/100 remaining
OpenRouter         -                        -      $0.00 spent
Vertex AI          Monthly                  -      0 tokens used
```

---

## Install

### Option 1 — Shell script (macOS & Linux, no dependencies)

```bash
curl -fsSL https://raw.githubusercontent.com/JValdivia23/quota-cli/main/install.sh | sh
```

Installs the latest release binary to `/usr/local/bin`. Override with:
```bash
INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/JValdivia23/quota-cli/main/install.sh | sh
```

### Option 2 — Homebrew (macOS & Linux)

```bash
brew tap JValdivia23/opencode-bar-curated
brew install qcli
```

### Option 3 — `go install` (requires Go 1.21+)

```bash
go install github.com/JValdivia23/quota-cli@latest
```

Make sure `$(go env GOPATH)/bin` is in your `$PATH`:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"   # add to ~/.zshrc or ~/.bashrc
```

### Option 4 — Download binary from GitHub Releases

Download the right archive for your platform from [GitHub Releases](https://github.com/JValdivia23/quota-cli/releases), extract, and move the `qcli` binary somewhere on your PATH.

| Platform | Archive |
|---|---|
| macOS (Apple Silicon) | `qcli_Darwin_arm64.tar.gz` |
| macOS (Intel) | `qcli_Darwin_x86_64.tar.gz` |
| Linux (x86_64) | `qcli_Linux_x86_64.tar.gz` |
| Windows | `qcli_Windows_x86_64.zip` |

### Option 5 — Build from source

```bash
git clone https://github.com/JValdivia23/quota-cli.git
cd quota-cli
go build -o qcli main.go
sudo mv qcli /usr/local/bin/   # or: mv qcli ~/bin/
```

---

## Usage

```bash
qcli              # Show all detected providers (same as qcli status)
qcli status       # Show quota and cost for every discovered provider
qcli status -j    # Output as JSON
qcli list         # List which providers were auto-detected
qcli report       # Deep-dive with 7-day trend and monthly forecast
```

---

## Credential discovery (zero config)

`qcli` searches for credentials automatically — no `.env` file required:

| Source | What it provides |
|---|---|
| `~/.local/share/opencode/auth.json` | OpenAI, OpenRouter, Copilot, Gemini |
| `antigravity-accounts.json` | Claude + Gemini OAuth tokens |
| `$OPENAI_API_KEY` | OpenAI |
| `$ANTHROPIC_API_KEY` | Claude |
| `$OPENROUTER_API_KEY` | OpenRouter |
| `$GITHUB_TOKEN` / `$COPILOT_TOKEN` | GitHub Copilot |
| `$GEMINI_API_KEY` | Gemini |
| `opencode.db` (SQLite) | OpenCode Zen |
| Google Application Default Credentials | Vertex AI |

---

## Project structure

```
cmd/quota/       CLI commands (status, list, report)
pkg/providers/   One file per AI provider, auto-registered
pkg/auth/        Credential discovery (auth.json, env vars, SQLite)
pkg/display/     Adaptive table + JSON output
pkg/predictor/   Weighted 7-day forecast engine
pkg/models/      Shared data types
```

---

## Development

```bash
go run main.go status      # Run locally
go build -o qcli main.go   # Build binary
go fmt ./...               # Format
go test ./...              # Test
```

To cut a release (triggers GoReleaser + Homebrew tap update):
```bash
git tag v1.2.3
git push origin v1.2.3
```

## License

[MIT](LICENSE)
