package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("4")).
			Foreground(lipgloss.Color("0")).
			Padding(0, 2)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("7")).
			Background(lipgloss.Color("236")).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Padding(0, 2)

	statusStyle = lipgloss.NewStyle()

	alignStyle = func(s string, width int) string {
		if len(s) > width {
			return s[:width]
		}
		return s + strings.Repeat(" ", width-len(s))
	}

	centeredStyle = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Padding(2)

	detailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("7")).
			Width(100).
			Align(lipgloss.Center).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2"))

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))

	diffAddedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2"))

	diffRemovedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))
)
