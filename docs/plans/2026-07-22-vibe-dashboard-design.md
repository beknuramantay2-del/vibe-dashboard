# vibe-dashboard: Design Document

## Overview

vibe-dashboard is a cross-platform **desktop application** that provides real-time monitoring and control of AI coding agents (Claude Code, OpenCode, Codex CLI). It displays active sessions, token/cost burn rates, cache hit rates, file diffs, and supports rollback — all in a native desktop window.

## Architecture

### Stack
- **Backend:** Go 1.25+
- **Desktop Framework:** Wails v2 (Go backend + webview frontend)
- **Frontend:** Svelte 5 + Vite
- **UI Fonts:** Geist Sans + Geist Mono
- **Database:** SQLite via modernc.org/sqlite (pure Go, no CGO)

### Data Sources

| Agent | Data Source | Detection |
|-------|-----------|-----------|
| Claude Code | `~/.claude/projects/**/*.jsonl` | JSONL parser with file glob |
| OpenCode | `~/.opencode/opencode.db` | SQLite read-only connection |
| Codex CLI | `~/.codex/logs/**/*.jsonl` | JSONL parser with file glob |

Each source implements the `SourceReader` interface with `Refresh()`, `ListSessions()`, `GetSession()`, `GetFileChanges()`, and `KillSession()`.

### Local Store

Aggregated data lives in `~/.vibe-dashboard/vibe.db` (SQLite). This enables:
- History across sessions
- Per-project and per-agent cost aggregation
- Budget tracking over time

### Desktop Layout

```
┌──────────┬──────────────────────────────────────────┐
│ Sidebar  │  Main Content                             │
│          │                                           │
│ Sessions │  Tab: Sessions — sortable table           │
│ Detail   │  Tab: Detail — stats, file changes, kill  │
│ Diff     │  Tab: Diff — before/after text compare    │
│ Snapshots│  Tab: Snapshots — rollback management     │
│ Config   │  Tab: Config — theme, agents, about       │
│          │                                           │
│ Agents   │                                           │
│ Cost     │                                           │
│ ⟳  ☀/☾  │                                           │
└──────────┴──────────────────────────────────────────┘
```

### Key Features

- **Real-time polling**: 3-second refresh cycle with Wails EventsEmit
- **Multi-agent support**: unified view of all AI coding sessions
- **Session sorting**: by agent, cost, tokens, cache hit rate, time
- **Cache hit rate**: color-coded (green ≥80%, amber ≥50%, red <50%)
- **Diff viewer**: line-by-line unified diff with line numbers
- **Rollback**: git stash-based snapshot create/restore/delete
- **Toast notifications**: success/error/info feedback
- **Theme toggle**: dark/light with localStorage persistence
- **Cross-platform**: Windows, macOS, Linux via Wails

### Error Handling

- Source adapters auto-detect availability; graceful skip if unavailable
- All frontend-facing methods return `ResultDTO { ok, message }` for error feedback
- Toast notifications surface errors to the user
- Backend logs to `~/.vibe-dashboard/vibe-desktop.log`
- Proper context cancellation on shutdown

### Security

- Repo paths validated: absolute, resolved symlinks, verified as git repos
- PID ownership verified before sending signals
- Snapshot IDs sanitized to prevent path traversal
- SQLite queries use parameterized statements
- File size limits (100MB) and entry limits (100K) on JSONL parsing

## Implementation Plan

### v0.1 (Complete)
- Project setup, Wails scaffold
- Source readers for Claude Code + OpenCode + Codex CLI
- Store layer (SQLite)
- Basic desktop UI: sessions list + detail view

### v0.2 (Current)
- Fixed critical bugs (readers never refreshing, duration calculation, cache rate)
- Security hardening (path validation, PID checks, snapshot ID sanitization)
- Frontend: sorting, search, toast notifications, loading states
- Snapshot management panel
- Proper shutdown (context cancellation, store close, log file close)
- Theme persistence

### v0.3 (Planned)
- Budget limits with configurable thresholds
- Report export (Markdown/CSV)
- Per-project cost aggregation views
- Keyboard shortcuts
- System tray integration
