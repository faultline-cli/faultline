package renderer

import lipgloss "charm.land/lipgloss/v2"

type styles struct {
	title      lipgloss.Style
	subtitle   lipgloss.Style
	muted      lipgloss.Style
	divider    lipgloss.Style
	card       lipgloss.Style
	callout    lipgloss.Style
	metaLabel  lipgloss.Style
	severity   map[string]lipgloss.Style
	confidence lipgloss.Style
}

func newStyles() styles {
	baseCard := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5B6470")).
		Padding(0, 1)

	return styles{
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F5F7FA")),
		subtitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#D6D9DE")),
		muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9AA4B2")),
		divider: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#697586")),
		card:       baseCard,
		callout:    baseCard.BorderForeground(lipgloss.Color("#7C8798")),
		metaLabel:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AEB7C2")),
		confidence: lipgloss.NewStyle().Foreground(lipgloss.Color("#8DD0A5")).Bold(true),
		severity: map[string]lipgloss.Style{
			"critical": pillStyle("#7F1D1D", "#FECACA"),
			"high":     pillStyle("#7C2D12", "#FED7AA"),
			"medium":   pillStyle("#713F12", "#FEF3C7"),
			"low":      pillStyle("#1E3A5F", "#BFDBFE"),
			"unknown":  pillStyle("#334155", "#E2E8F0"),
		},
	}
}

func pillStyle(bg, fg string) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(fg)).
		Padding(0, 1)
}
