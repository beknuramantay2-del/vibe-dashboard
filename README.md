# vibe-dashboard

Desktop dashboard for monitoring AI coding agents (Claude Code, OpenCode, Codex CLI).

Track sessions, tokens, costs, cache hit rates, file diffs, and rollback changes in a native desktop app.

## Features

- Multi-agent monitoring (Claude Code + OpenCode + Codex CLI in one window)
- Real-time burn rate and cost tracking
- Cache hit rate visualization (color-coded)
- Session list with active/recent history
- Diff viewer for changed files
- One-key rollback via git stash
- Dark/light theme toggle
- Cross-platform (Windows, macOS, Linux)

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

# Binary is at build/bin/vibe-desktop.exe (Windows)
# or build/bin/vibe-dashboard (macOS/Linux)
```

## Development

```bash
# Run in development mode with hot-reload
wails dev
```

## Data Sources

| Agent | Data Source | Detection |
|-------|------------|-----------|
| Claude Code | `~/.claude/projects/**/*.jsonl` | Auto |
| OpenCode | `~/.opencode/opencode.db` | Auto |
| Codex CLI | `~/.codex/logs/**/*.jsonl` | Auto |

## Architecture

```
vibe-dashboard/
├── main.go               # Wails entry point
├── app.go                # Backend bindings (sessions, cost, kill, rollback)
├── sources/              # Agent data readers
│   ├── claude.go         # Claude Code JSONL parser
│   ├── opencode.go       # OpenCode SQLite reader
│   └── codex.go          # Codex CLI JSONL parser
├── frontend/             # Svelte desktop UI
│   └── src/
│       ├── App.svelte         # Root layout
│       └── lib/
│           ├── Sidebar.svelte       # Navigation + agent list
│           ├── SessionList.svelte   # Session table with filters
│           ├── SessionDetail.svelte # Session info + actions
│           ├── DiffViewer.svelte    # Before/after diff
│           └── ConfigPanel.svelte   # Theme + agent info
├── store/                # Local SQLite store
├── rollback/             # Git-based rollback
└── docs/plans/           # Design documents
```

## License

MIT
