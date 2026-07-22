package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type BudgetAlert struct {
	Name       string
	Threshold  float64
	Action     string
	Current    float64
}

type CostTracker struct {
	totalCost    float64
	totalInput   int64
	totalOutput  int64
	cacheRate    float64
	alerts       []BudgetAlert
}

func NewCostTracker() CostTracker {
	return CostTracker{
		alerts: []BudgetAlert{
			{Name: "Daily Soft Cap", Threshold: 5.00, Action: "warn"},
			{Name: "Daily Hard Cap", Threshold: 10.00, Action: "stop"},
			{Name: "Session Limit", Threshold: 2.00, Action: "warn"},
		},
	}
}

func (ct *CostTracker) Render(totalCost, totalInput, totalOutput, cacheRate float64) string {
	ct.totalCost = totalCost
	ct.totalInput = int64(totalInput)
	ct.totalOutput = int64(totalOutput)
	ct.cacheRate = cacheRate

	var b strings.Builder

	b.WriteString(sectionStyle.Render("COST OVERVIEW") + "\n\n")

	ct.renderMetric(&b, "Total Cost", fmt.Sprintf("$%.2f", totalCost))
	ct.renderMetric(&b, "Input Tokens", fmt.Sprintf("%.0f", totalInput))
	ct.renderMetric(&b, "Output Tokens", fmt.Sprintf("%.0f", totalOutput))
	ct.renderMetric(&b, "Cache Hit Rate", fmt.Sprintf("%.1f%%", cacheRate))

	if cacheRate > 80 {
		b.WriteString(infoStyle.Render("  ✓ Cache usage is excellent\n"))
	} else if cacheRate > 50 {
		b.WriteString(warnStyle.Render("  ⚠ Cache could be improved\n"))
	} else {
		b.WriteString(errorStyle.Render("  ✗ Low cache hit rate — enable caching!\n"))
	}

	b.WriteString("\n" + sectionStyle.Render("BURN RATE") + "\n\n")
	b.WriteString(fmt.Sprintf("  Rate: $%.2f/min\n", totalCost/60))
	b.WriteString(fmt.Sprintf("  Projected (1h): $%.2f\n", totalCost))

	b.WriteString("\n" + sectionStyle.Render("BUDGET ALERTS") + "\n\n")
	for _, a := range ct.alerts {
		color := lipgloss.Color("2")
		if a.Current >= a.Threshold {
			color = lipgloss.Color("1")
		} else if a.Current >= a.Threshold*0.7 {
			color = lipgloss.Color("3")
		}

		bar := ct.renderBar(a.Current, a.Threshold, 20)
		b.WriteString(fmt.Sprintf("  %s %s %s %s $%.2f / $%.2f\n",
			statusStyle.Copy().Foreground(color).Render("●"),
			padRight(a.Name, 18),
			bar,
			padRight(a.Action, 6),
			a.Current, a.Threshold,
		))
	}

	return b.String()
}

func (ct *CostTracker) renderMetric(b *strings.Builder, name, value string) {
	b.WriteString(fmt.Sprintf("  %s: %s\n", padRight(name+":", 16), value))
}

func (ct *CostTracker) renderBar(current, threshold float64, width int) string {
	ratio := current / threshold
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	color := lipgloss.Color("2")
	if ratio > 0.85 {
		color = lipgloss.Color("1")
	} else if ratio > 0.6 {
		color = lipgloss.Color("3")
	}

	return lipgloss.NewStyle().Foreground(color).Render(bar)
}

func (ct *CostTracker) SetAlertCurrent(name string, current float64) {
	for i, a := range ct.alerts {
		if a.Name == name {
			ct.alerts[i].Current = current
			return
		}
	}
}
