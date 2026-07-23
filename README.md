# vibe-dashboard

Desktop dashboard for monitoring AI coding agents (Claude Code, OpenCode, Codex CLI).

Track sessions, tokens, costs, cache hit rates, file diffs, and rollback changes in a native desktop app.

## Features

- **Multi-agent monitoring** — Claude Code + OpenCode + Codex CLI in one window
- **Real-time updates** — 3-second polling with live session status
- **Cost tracking** — Per-session and aggregated cost with burn rate
- **Cache hit rate** — Color-coded visualization (green/amber/red)
- **Session management** — Search, filter (active/done), sort by any column
- **Session kill** — Terminate active agent processes with one click
- **Diff viewer** — Before/after text comparison with line numbers
- **Rollback** — Git stash-based snapshots: create, restore, delete
- **Dark/light theme** — Toggle with localStorage persistence
- **Cross-platform** — Windows, macOS, Linux

## Quick Start

### Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.25+ | https://go.dev/dl/ |
| Node.js | 20+ | https://nodejs.org/ |
| Wails CLI | v2 | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |

**Linux only** — install system dependencies first:
```bash
# Ubuntu/Debian
sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install gtk3-devel webkit2gtk4.0-devel

# Arch
sudo pacman -S gtk3 webkit2gtk-4.1
```

**macOS** — Xcode Command Line Tools required:
```bash
xcode-select --install
```

**Windows** — WebView2 runtime (pre-installed on Windows 11, download for Windows 10 from Microsoft).

### Verify setup

```bash
wails doctor
```

This checks all dependencies are installed. Fix any issues it reports before building.

### Build & run

```bash
git clone https://github.com/beknuramantay2-del/vibe-dashboard.git
cd vibe-dashboard

# Install frontend dependencies + build + compile Go → single binary
wails build

# Run the binary
./build/bin/vibe-wails          # Linux/macOS
.\build\bin\vibe-wails.exe      # Windows
```

### Development mode (hot-reload)

```bash
wails dev
```

Opens a desktop window with live-reloading frontend. Backend changes require restart.

## Data Sources

The app auto-detects installed agents. No configuration needed.

| Agent | Data Source | Detection |
|-------|-----------|-----------|
| Claude Code | `~/.claude/projects/**/*.jsonl` | Auto |
| OpenCode | `~/.opencode/opencode.db` | Auto |
| Codex CLI | `~/.codex/logs/**/*.jsonl` | Auto |

If no agents are installed, the dashboard opens with an empty session list — install at least one agent to see data.

## Architecture

```
vibe-dashboard/
├── main.go                    # Wails entry point
├── app.go                     # Backend bindings (sessions, cost, kill, rollback)
├── sources/                   # Agent data readers
│   ├── interface.go           # SourceReader interface + Session/FileChange types
│   ├── claude.go              # Claude Code JSONL parser
│   ├── opencode.go            # OpenCode SQLite reader
│   ├── codex.go               # Codex CLI JSONL parser
│   ├── helpers.go             # Shared utilities (WalkDir, killProcess)
│   ├── signal_unix.go         # Unix-specific signal handling
│   └── signal_windows.go      # Windows-specific signal handling
├── store/                     # Local SQLite store for aggregation
│   └── db.go
├── rollback/                  # Git stash-based snapshot system
│   └── snapshot.go
├── frontend/                  # Svelte desktop UI
│   └── src/
│       ├── App.svelte         # Root layout + state management
│       ├── main.js            # Svelte mount entry point
│       └── lib/               # UI components
└── wails.json                 # Wails build configuration
```

## Storage

All data is stored locally in `~/.vibe-dashboard/`:

| File | Purpose |
|------|---------|
| `vibe.db` | Aggregated session data (SQLite) |
| `snapshots/` | Rollback snapshot metadata |
| `vibe-desktop.log` | Application logs (auto-rotated at 10MB) |

## Troubleshooting

**`wails: command not found`**
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
# Make sure $GOPATH/bin is in your PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

**`wails doctor` reports missing dependencies**
Install the system packages listed above for your OS.

**Blank window / no sessions**
- Check `~/.vibe-dashboard/vibe-desktop.log` for errors
- Ensure at least one agent (Claude Code, OpenCode, or Codex CLI) has been used — the app reads their log files

**Build fails with embed error**
```bash
# The frontend must be built first for go:embed to work
cd frontend && npm install && npm run build && cd ..
wails build
```

## License

MIT
