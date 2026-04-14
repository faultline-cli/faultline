package fixtures

import (
	"strings"

	"golang.org/x/net/html"
)

func htmlToText(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	var b strings.Builder
	z := html.NewTokenizer(strings.NewReader(input))
	lastWasSpace := false
	lastWasNewline := false
	inPre := false

	emit := func(text string) {
		text = html.UnescapeString(text)
		if !inPre {
			text = strings.Join(strings.Fields(text), " ")
		}
		if text == "" {
			return
		}
		if inPre {
			b.WriteString(text)
			lastWasSpace = false
			lastWasNewline = strings.HasSuffix(text, "\n")
			return
		}
		if lastWasNewline {
			b.WriteString(text)
		} else if lastWasSpace {
			b.WriteByte(' ')
			b.WriteString(text)
		} else if b.Len() > 0 {
			b.WriteByte(' ')
			b.WriteString(text)
		} else {
			b.WriteString(text)
		}
		lastWasSpace = false
		lastWasNewline = false
	}

	writeNewline := func() {
		if b.Len() == 0 {
			return
		}
		if !strings.HasSuffix(b.String(), "\n") {
			b.WriteByte('\n')
		}
		lastWasSpace = false
		lastWasNewline = true
	}

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return cleanTextOutput(b.String())
		case html.TextToken:
			emit(string(z.Text()))
		case html.StartTagToken, html.SelfClosingTagToken:
			name, hasAttr := z.TagName()
			_ = hasAttr
			tag := strings.ToLower(string(name))
			switch tag {
			case "br", "p", "div", "section", "article", "blockquote", "li", "tr", "table", "thead", "tbody", "tfoot", "ul", "ol":
				writeNewline()
			case "pre":
				inPre = true
				writeNewline()
			}
		case html.EndTagToken:
			name, _ := z.TagName()
			tag := strings.ToLower(string(name))
			switch tag {
			case "p", "div", "section", "article", "blockquote", "li", "tr", "table", "thead", "tbody", "tfoot", "ul", "ol":
				writeNewline()
			case "pre":
				inPre = false
				writeNewline()
			}
		}
	}
}

func cleanTextOutput(text string) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(lines))
	seenBlank := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if !seenBlank {
				out = append(out, "")
				seenBlank = true
			}
			continue
		}
		out = append(out, line)
		seenBlank = false
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
