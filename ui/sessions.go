package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/vibe-dashboard/vibe-dashboard/sources"
)

type SessionsPane struct {
	viewport   viewport.Model
	sessions   []sources.Session
	selected   int
	expanded   bool
}

func NewSessionsPane(width, height int) SessionsPane {
	return SessionsPane{
		viewport: viewport.New(width, height),
	}
}

func (sp *SessionsPane) SetSessions(s []sources.Session) {
	sp.sessions = s
	sp.updateContent()
}

func (sp *SessionsPane) updateContent() {
	var b strings.Builder

	for i, s := range sp.sessions {
		prefix := "  "
		if i == sp.selected {
			prefix = " >"
		}

		duration := formatDuration(s.Duration)
		agentColor := agentColor(s.Agent)

		statusDot := "●"
		dotColor := lipgloss.Color("2")
		if s.Status == "active" {
			dotColor = lipgloss.Color("3")
		} else if s.Status == "completed" {
			dotColor = lipgloss.Color("8")
		}

		b.WriteString(fmt.Sprintf("%s %s %s %s %s %s $%.2f\n",
			prefix,
			statusStyle.Copy().Foreground(dotColor).Render(statusDot),
			statusStyle.Copy().Foreground(agentColor).Render(padRight(s.Agent, 8)),
			padRight(s.Project, 18),
			duration,
			padRight(s.Status, 10),
			s.Cost,
		))

		if i == sp.selected && sp.expanded {
			b.WriteString(sp.renderDetail(s))
		}
	}

	sp.viewport.SetContent(lipgloss.NewStyle().Padding(0, 2).Render(b.String()))
}

func (sp *SessionsPane) renderDetail(s sources.Session) string {
	var d strings.Builder
	d.WriteString(fmt.Sprintf("  ID: %s\n", s.ID))
	d.WriteString(fmt.Sprintf("  Input:  %d tokens\n", s.InputTokens))
	d.WriteString(fmt.Sprintf("  Output: %d tokens\n", s.OutputTokens))
	d.WriteString(fmt.Sprintf("  Cache:  %d tokens (%.1f%% hit rate)\n", s.CacheTokens, s.CacheHitRate))
	d.WriteString(fmt.Sprintf("  Cost:   $%.4f\n", s.Cost))
	return detailStyle.Render(d.String()) + "\n"
}

func (sp *SessionsPane) View() string {
	return sp.viewport.View()
}

func (sp *SessionsPane) Update(msg tea.Msg) (SessionsPane, tea.Cmd) {
	var cmd tea.Cmd
	sp.viewport, cmd = sp.viewport.Update(msg)
	return *sp, cmd
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func agentColor(agent string) lipgloss.Color {
	switch agent {
	case "claude":
		return lipgloss.Color("4")
	case "opencode":
		return lipgloss.Color("5")
	case "codex":
		return lipgloss.Color("6")
	default:
		return lipgloss.Color("7")
	}
}

func padRight(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s + strings.Repeat(" ", n-len(s))
}
