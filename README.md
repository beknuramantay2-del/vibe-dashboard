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

## Screenshots

_Coming soon_

## Installation

### Download

Download the latest release from the [releases page](https://github.com/beknuramantay2-del/vibe-dashboard/releases).

### Build from source

Prerequisites: Go 1.25+, Node.js 20+

```bash
git clone https://github.com/beknuramantay2-del/vibe-dashboard.git
cd vibe-dashboard

# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Build
wails build

# Binary is at build/bin/vibe-wails.exe (Windows)
# or build/bin/vibe-wails (macOS/Linux)
```

## Development

```bash
# Run in development mode with hot-reload
wails dev
```

## Data Sources

| Agent | Data Source | Detection |
|-------|-----------|-----------|
| Claude Code | `~/.claude/projects/**/*.jsonl` | Auto |
| OpenCode | `~/.opencode/opencode.db` | Auto |
| Codex CLI | `~/.codex/logs/**/*.jsonl` | Auto |

## Architecture

```
vibe-dashboard/
├── main.go                    # Wails entry point
├── app.go                     # Backend bindings (sessions, cost, kill, rollback)
├── sources/                   # Agent data readers
│   ├── interface.go           # SourceReader interface + Session/FileChange types
│   ├── claude.go              # Claude Code JSONL parser
│   ├── opencode.go            # OpenCode SQLite reader
│   └── codex.go               # Codex CLI JSONL parser
├── store/                     # Local SQLite store for aggregation
│   └── db.go
├── rollback/                  # Git stash-based snapshot system
│   └── snapshot.go
├── frontend/                  # Svelte desktop UI
│   └── src/
│       ├── App.svelte              # Root layout + state management
│       ├── style.css               # Global styles + theme variables
│       └── lib/
│           ├── Sidebar.svelte      # Navigation + agent list + cost
│           ├── SessionList.svelte  # Sortable session table with search
│           ├── SessionDetail.svelte # Stats, file changes, kill, snapshot
│           ├── DiffViewer.svelte   # Before/after text diff
│           ├── SnapshotPanel.svelte # Rollback snapshot management
│           ├── ConfigPanel.svelte  # Theme, agents, data sources
│           └── Toast.svelte        # Toast notification system
├── docs/plans/                # Design documents
├── wails.json                 # Wails configuration
├── go.mod                     # Go module
└── Makefile                   # Build commands
```

## Storage

All data is stored locally at `~/.vibe-dashboard/`:
- `vibe.db` — Aggregated session database
- `snapshots/` — Rollback snapshot metadata
- `vibe-desktop.log` — Application logs

## License

MIT
