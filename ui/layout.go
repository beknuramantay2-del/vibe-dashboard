package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/vibe-dashboard/vibe-dashboard/sources"
	"github.com/vibe-dashboard/vibe-dashboard/store"
)

type Tab int

const (
	TabSessions Tab = iota
	TabTokens
	TabDiff
	TabConfig
)

type Model struct {
	store        *store.Store
	readers      []sources.SourceReader
	sessions     []sources.Session
	selectedIdx  int
	activeTab    Tab
	width        int
	height       int
	ready        bool
	sessionsPane viewport.Model
	diffPane     viewport.Model
	err          error
}

func NewModel(store *store.Store, readers []sources.SourceReader) Model {
	return Model{
		store:   store,
		readers: readers,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.sessionsPane = viewport.New(msg.Width-4, msg.Height-10)
			m.sessionsPane.YPosition = 3
			m.diffPane = viewport.New(msg.Width-4, 8)
			m.ready = true
		} else {
			m.sessionsPane.Width = msg.Width - 4
			m.sessionsPane.Height = msg.Height - 10
			m.diffPane.Width = msg.Width - 4
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % 4
		case "shift+tab":
			m.activeTab = (m.activeTab - 1 + 4) % 4
		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case "down", "j":
			if m.selectedIdx < len(m.sessions)-1 {
				m.selectedIdx++
			}
		case "r":
			ses, err := m.selectedSession()
			if err == nil {
				cmds = append(cmds, func() tea.Msg {
					return RollbackMsg{SessionID: ses.ID, Project: ses.Project}
				})
			}
		}

		var cmd tea.Cmd
		m.sessionsPane, cmd = m.sessionsPane.Update(msg)
		cmds = append(cmds, cmd)
		m.diffPane, cmd = m.diffPane.Update(msg)
		cmds = append(cmds, cmd)

	case SessionsUpdatedMsg:
		m.sessions = msg.Sessions
		m.sessionsPane.SetContent(m.renderSessions())

	case RollbackResultMsg:
		if msg.Err != nil {
			m.err = msg.Err
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	header := m.renderHeader()
	body := m.renderBody()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		body,
		footer,
	)
}

func (m Model) renderHeader() string {
	tabs := []string{"Sessions", "Tokens/Cost", "Diff", "Config"}
	var renderedTabs []string
	for i, t := range tabs {
		if Tab(i) == m.activeTab {
			renderedTabs = append(renderedTabs, activeTabStyle.Render(t))
		} else {
			renderedTabs = append(renderedTabs, inactiveTabStyle.Render(t))
		}
	}

	title := titleStyle.Render(" vibe-dashboard v0.1 ")
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	return lipgloss.JoinHorizontal(lipgloss.Center, title, tabBar)
}

func (m Model) renderBody() string {
	switch m.activeTab {
	case TabSessions:
		return m.sessionsPane.View()
	case TabTokens:
		return m.renderTokensCost()
	case TabDiff:
		return m.diffPane.View()
	case TabConfig:
		return m.renderConfig()
	default:
		return "Unknown tab"
	}
}

func (m Model) renderFooter() string {
	return footerStyle.Render(
		" [↑↓] nav  [Enter] details  [r] rollback  [Tab] switch  [q] quit ",
	)
}

func (m Model) selectedSession() (*sources.Session, error) {
	if m.selectedIdx < 0 || m.selectedIdx >= len(m.sessions) {
		return nil, fmt.Errorf("no session selected")
	}
	return &m.sessions[m.selectedIdx], nil
}

func (m Model) renderSessions() string {
	if len(m.sessions) == 0 {
		return centeredStyle.Render("No active sessions\n\nWaiting for agents...")
	}

	var b strings.Builder
	for i, s := range m.sessions {
		prefix := " "
		if i == m.selectedIdx {
			prefix = ">"
		}

		statusColor := lipgloss.Color("2")
		if s.Status == "active" {
			statusColor = lipgloss.Color("3")
		} else if s.Status == "completed" {
			statusColor = lipgloss.Color("8")
		}

		duration := fmt.Sprintf("%dm", int(s.Duration.Minutes()))
		if s.Duration.Hours() >= 1 {
			duration = fmt.Sprintf("%dh%02dm", int(s.Duration.Hours()), int(s.Duration.Minutes())%60)
		}

		line := fmt.Sprintf("%s %s %s  %s · %s  $%.2f",
			prefix,
			statusStyle.Copy().Foreground(statusColor).Render("●"),
			alignStyle(s.Agent, 8),
			alignStyle(s.Project, 20),
			duration,
			s.Cost,
		)
		b.WriteString(line + "\n")

		if i == m.selectedIdx {
			b.WriteString(renderSessionDetail(s) + "\n")
		}
	}
	return b.String()
}

func renderSessionDetail(s sources.Session) string {
	total := s.InputTokens + s.OutputTokens
	var tokensStr string
	if total > 1000 {
		tokensStr = fmt.Sprintf("%dk", total/1000)
	} else {
		tokensStr = fmt.Sprintf("%d", total)
	}

	cacheColor := lipgloss.Color("2")
	if s.CacheHitRate < 50 {
		cacheColor = lipgloss.Color("1")
	} else if s.CacheHitRate < 80 {
		cacheColor = lipgloss.Color("3")
	}

	return detailStyle.Render(fmt.Sprintf(
		"    Input: %d tok  Output: %d tok  Cache: %.0f%% %s  $%.2f",
		s.InputTokens, s.OutputTokens, s.CacheHitRate,
		statusStyle.Copy().Foreground(cacheColor).Render("●"),
		s.Cost,
	))
}

func (m Model) renderTokensCost() string {
	var totalCost, totalInput, totalOutput float64
	for _, s := range m.sessions {
		totalCost += s.Cost
		totalInput += float64(s.InputTokens)
		totalOutput += float64(s.OutputTokens)
	}

	var b strings.Builder
	b.WriteString(sectionStyle.Render("TOKEN & COST OVERVIEW") + "\n\n")

	b.WriteString(fmt.Sprintf("Total Cost:       $%.2f\n", totalCost))
	b.WriteString(fmt.Sprintf("Input Tokens:     %.0fk\n", totalInput/1000))
	b.WriteString(fmt.Sprintf("Output Tokens:    %.0fk\n", totalOutput/1000))

	if totalInput+totalOutput > 0 {
		cacheRate := 0.0
		for _, s := range m.sessions {
			cacheRate += s.CacheHitRate
		}
		cacheRate /= float64(len(m.sessions))
		b.WriteString(fmt.Sprintf("Cache Hit Rate:   %.1f%%\n", cacheRate))
	}

	b.WriteString("\n" + sectionStyle.Render("SESSION COSTS") + "\n\n")
	for _, s := range m.sessions {
		b.WriteString(fmt.Sprintf("  %s/%s  $%.2f  (%dm)\n",
			s.Agent, s.Project, s.Cost, int(s.Duration.Minutes())))
	}

	return b.String()
}

func (m Model) renderConfig() string {
	var b strings.Builder
	b.WriteString(sectionStyle.Render("CONFIGURATION") + "\n\n")
	b.WriteString("  [b] Set budget alert\n")
	b.WriteString("  [e] Export report (CSV)\n")
	b.WriteString("  [m] Toggle team mode\n")
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("BUDGET ALERTS") + "\n")
	b.WriteString("  No alerts configured. Press [b] to add one.\n")
	return b.String()
}

type SessionsUpdatedMsg struct {
	Sessions []sources.Session
}

type RollbackMsg struct {
	SessionID string
	Project   string
}

type RollbackResultMsg struct {
	Err error
}
