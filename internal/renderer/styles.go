package renderer

import (
	ansi "charm.land/glamour/v2/ansi"
	lipgloss "charm.land/lipgloss/v2"
)

type styles struct {
	title      lipgloss.Style
	subtitle   lipgloss.Style
	muted      lipgloss.Style
	divider    lipgloss.Style
	card       lipgloss.Style
	callout    lipgloss.Style
	panel      lipgloss.Style
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
		panel:      baseCard.Padding(0, 1),
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

func panelTitleStyle(bg, fg string) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(fg)).
		Padding(0, 1)
}

func markdownStyles(dark bool) ansi.StyleConfig {
	body := "#1F2933"
	muted := "#52606D"
	accent := "#102A43"
	link := "#1D4ED8"
	codeFg := "#0F172A"
	codeBg := "#E5EAF1"
	separator := "#C0CAD5"
	if dark {
		body = "#E5E7EB"
		muted = "#AEB7C2"
		accent = "#F5F7FA"
		link = "#8CB7FF"
		codeFg = "#F5F7FA"
		codeBg = "#24313D"
		separator = "#5B6470"
	}

	zero := uint(0)
	one := uint(1)
	two := uint(2)
	bullet := "• "
	pipe := "│ "
	column := " │ "
	row := "─"
	empty := ""
	numberSuffix := ". "
	bold := true
	italic := true
	underline := true

	return ansi.StyleConfig{
		Document:  ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{BlockPrefix: "", BlockSuffix: ""}, Margin: &zero},
		Paragraph: ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: &body}, Margin: &zero},
		List: ansi.StyleList{
			StyleBlock:  ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: &body}, Margin: &zero},
			LevelIndent: 2,
		},
		Item:        ansi.StylePrimitive{Color: &body, Prefix: bullet},
		Enumeration: ansi.StylePrimitive{Color: &body, Bold: &bold, Suffix: numberSuffix},
		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{Color: &muted, BlockPrefix: pipe, BlockSuffix: ""},
			Indent:         &zero,
			Margin:         &zero,
		},
		Heading:  ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: &accent, Bold: &bold}, Margin: &zero},
		H1:       ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: &accent, Bold: &bold, BlockPrefix: "", BlockSuffix: ""}, Margin: &one},
		H2:       ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: &accent, Bold: &bold, BlockPrefix: "", BlockSuffix: ""}, Margin: &one},
		H3:       ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: &accent, Bold: &bold, BlockPrefix: "", BlockSuffix: ""}, Margin: &one},
		H4:       ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: &accent, Bold: &bold}, Margin: &zero},
		H5:       ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: &accent, Bold: &bold}, Margin: &zero},
		H6:       ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: &accent, Bold: &bold}, Margin: &zero},
		Emph:     ansi.StylePrimitive{Color: &body, Italic: &italic},
		Strong:   ansi.StylePrimitive{Color: &body, Bold: &bold},
		Link:     ansi.StylePrimitive{Color: &link, Underline: &underline},
		LinkText: ansi.StylePrimitive{Color: &link, Underline: &underline},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{Color: &codeFg, BackgroundColor: &codeBg},
			Margin:         &zero,
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{Color: &codeFg, BackgroundColor: &codeBg, BlockPrefix: "", BlockSuffix: ""},
				Indent:         &two,
				Margin:         &zero,
			},
		},
		Table: ansi.StyleTable{
			StyleBlock:      ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: &body}, Margin: &zero},
			CenterSeparator: &empty,
			ColumnSeparator: &column,
			RowSeparator:    &row,
		},
		HorizontalRule:        ansi.StylePrimitive{Color: &separator},
		DefinitionList:        ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: &body}, Margin: &zero},
		DefinitionTerm:        ansi.StylePrimitive{Color: &accent, Bold: &bold},
		DefinitionDescription: ansi.StylePrimitive{Color: &body},
		HTMLBlock:             ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: &body}, Margin: &zero},
		HTMLSpan:              ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: &body}, Margin: &zero},
		Task:                  ansi.StyleTask{StylePrimitive: ansi.StylePrimitive{Color: &body}, Ticked: "✓ ", Unticked: "○ "},
		Text:                  ansi.StylePrimitive{Color: &body},
	}
}
