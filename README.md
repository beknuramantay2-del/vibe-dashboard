# vibe-dashboard

Real-time TUI dashboard for monitoring AI coding agents (Claude Code, OpenCode, Codex CLI).

Track sessions, tokens, costs, cache hit rates, file diffs, and rollback changes inside your terminal.

## Features

- Multi-agent monitoring (Claude Code + OpenCode + Codex CLI in one UI)
- Real-time burn rate and cost tracking
- Cache hit rate visualization (color-coded)
- Session list with active/recent history
- Diff viewer for changed files
- One-key rollback via git stash
- Budget alerts with progress bars
- Cross-platform (Windows, macOS, Linux)

## Installation

### macOS / Linux

```bash
curl -sSL https://github.com/vibe-dashboard/vibe-dashboard/releases/latest/download/install.sh | sh
```

### Windows (PowerShell)

```powershell
iwr -useb https://github.com/vibe-dashboard/vibe-dashboard/releases/latest/download/install.ps1 | iex
```

### Go install

```bash
go install github.com/vibe-dashboard/vibe-dashboard/cmd/vibe-dashboard@latest
```

## Usage

```bash
vibe-dashboard
```

### Hotkeys

| Key | Action |
|-----|--------|
| `↑`/`↓` | Navigate sessions |
| `Tab` | Switch tabs |
| `Enter` | Expand session details |
| `d` | Show file diff |
| `r` | Rollback session changes |
| `k` | Kill active session |
| `b` | Set budget alert |
| `q` / `Ctrl+C` | Quit |

## Data Sources

| Agent | Data Source | Detection |
|-------|------------|-----------|
| Claude Code | `~/.claude/projects/**/*.jsonl` | Auto |
| OpenCode | `~/.opencode/opencode.db` | Auto |
| Codex CLI | `~/.codex/logs/**/*.jsonl` | Auto |

## Architecture

```
vibe-dashboard/
├── cmd/vibe-dashboard/  # Entry point
├── sources/             # Agent data readers
│   ├── claude.go        # Claude Code JSONL parser
│   ├── opencode.go      # OpenCode SQLite reader
│   └── codex.go         # Codex CLI JSONL parser
├── ui/                  # Bubble Tea TUI
│   ├── layout.go        # Main layout + tabs
│   ├── sessions.go      # Session list panel
│   ├── tokens.go        # Token/cost charts
│   ├── diff.go          # Diff viewer panel
│   └── cost.go          # Budget tracking
├── store/               # Local SQLite store
│   └── db.go            # Aggregation + history
├── rollback/            # Git-based rollback
│   └── snapshot.go      # Git stash manager
└── docs/plans/          # Design documents
```

## Building from source

```bash
git clone https://github.com/vibe-dashboard/vibe-dashboard.git
cd vibe-dashboard
go build -o vibe-dashboard ./cmd/vibe-dashboard
```

## License

MIT
