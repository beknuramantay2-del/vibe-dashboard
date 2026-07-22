# vibe-dashboard — Design Document

## Overview

vibe-dashboard is a cross-platform TUI (Terminal User Interface) that provides real-time monitoring and control of AI coding agents (Claude Code, OpenCode, Codex CLI). It displays active sessions, token/cost burn rates, file diffs, and supports rollback — all in one terminal dashboard.

## Architecture

### Stack
- **Language:** Go 1.22+
- **TUI Framework:** Bubble Tea (charm.sh/bubbletea)
- **UI Components:** Bubbles (charm.sh/bubbles) + Lip Gloss (charm.sh/lipgloss)
- **Charts:** ntcharts (charm.sh/x/ntcharts)
- **Database:** SQLite via modernc.org/sqlite (pure Go, no CGO)
- **File Watching:** fsnotify (github.com/fsnotify/fsnotify)
- **Diff Library:** github.com/sergi/go-diff

### Data Sources

1. **Claude Code** — JSONL logs at `~/.claude/projects/**/*.jsonl`
2. **OpenCode** — SQLite database at `~/.opencode/opencode.db` (or deprecated JSON at `~/.local/share/opencode/storage/session/global/*.json`)
3. **Codex CLI** — JSONL logs at `~/.codex/logs/**/*.jsonl`

Each source has a unified `SourceReader` interface.

### Local Store

Aggregated data lives in `~/.vibe-dashboard/vibe.db` (SQLite). This enables:
- History across sessions
- Per-project aggregation
- Budget tracking over time

### TUI Layout

```
┌─────────────────────────────────────────────────────────┐
│ Header: logo + tabs (sessions/tokens/diff/config)       │
├─────────────┬───────────────────┬───────────────────────┤
│ Sessions    │ Tokens / Cost     │ Files (changed)        │
│ (active +   │ (burn graph,      │ (modified files per    │
│  recent)    │  progress bar,    │  selected session)     │
│             │  cache hit rate)  │                        │
├─────────────┴───────────────────┴───────────────────────┤
│ Diff Viewer (bottom panel, shows file diff on [d])      │
├─────────────────────────────────────────────────────────┤
│ Status bar: hotkeys help                                │
└─────────────────────────────────────────────────────────┘
```

### Key Features

- **Real-time burn rate** — fsnotify on JSONL + 500ms polling on SQLite
- **Multi-agent support** — unified view of all AI coding sessions
- **Cache hit rate** — color-coded (green > 80%, amber > 50%, red < 50%)
- **Diff viewer** — side-by-side or unified diff of changed files
- **Rollback** — git-based snapshot restore (one-key undo)
- **Budget alerts** — customizable thresholds with color warnings
- **Cross-platform** — pure Go, SQLite without CGO, static binary

### Error Handling

- Source adapters auto-detect version (SQLite first, fallback to JSON)
- Graceful degradation if a source is unavailable
- All TUI state is managed via Bubble Tea's model-update pattern
- Logging to file via `tea.LogToFile()` (no stdout conflict)

### Testing Strategy

- Unit tests for each source parser (test fixtures in testdata/)
- Store tests with in-memory SQLite
- UI tests with Bubble Tea's testing utilities
- Integration tests using mock data streams

## Implementation Plan

### Phase 1 (v0.1) — Must Have
- Project setup, go.mod, CI
- Source readers for Claude Code + OpenCode + Codex
- Store layer (SQLite)
- Basic TUI: sessions list + token/cost view
- File watchers for real-time updates

### Phase 2 (v0.2) — Should Have
- Diff viewer with go-diff
- Rollback system
- Per-project history
- Session kill (SIGTERM)

### Phase 3 (v0.3) — Nice to Have
- Budget limits with auto-stop
- Report export (Markdown/CSV)
- Team mode (shared SQLite)
- MCP integration
