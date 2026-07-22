package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type DiffViewer struct {
	viewport viewport.Model
	content  string
	width    int
	height   int
}

func NewDiffViewer(width, height int) DiffViewer {
	vp := viewport.New(width, height)
	return DiffViewer{
		viewport: vp,
		width:    width,
		height:   height,
	}
}

func (d *DiffViewer) SetSize(width, height int) {
	d.width = width
	d.height = height
	d.viewport.Width = width
	d.viewport.Height = height
}

func (d *DiffViewer) ShowDiff(before, after string) {
	d.content = d.renderDiff(before, after)
	d.viewport.SetContent(d.content)
}

func (d DiffViewer) View() string {
	if d.content == "" {
		return centeredStyle.Render("No diff selected\n\nPress [d] on a changed file")
	}
	return d.viewport.View()
}

func (d *DiffViewer) renderDiff(before, after string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(before, after, true)
	diffs = dmp.DiffCleanupSemantic(diffs)

	var b strings.Builder
	for _, diff := range diffs {
		text := diff.Text
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			switch diff.Type {
			case diffmatchpatch.DiffInsert:
				b.WriteString(diffAddedStyle.Render("+ " + line) + "\n")
			case diffmatchpatch.DiffDelete:
				b.WriteString(diffRemovedStyle.Render("- " + line) + "\n")
			case diffmatchpatch.DiffEqual:
				b.WriteString("  " + line + "\n")
			}
		}
	}
	return b.String()
}

func (d *DiffViewer) Update(msg tea.Msg) (DiffViewer, tea.Cmd) {
	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return d, cmd
}
