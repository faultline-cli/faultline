package app

import (
	"context"
	"fmt"
	"io"
	"strings"

	"faultline/internal/store"
)

type reportJSON struct {
	Store    historyStoreJSON      `json:"store"`
	Failures []store.FailureReport `json:"failures"`
}

func (Service) Report(ctx context.Context, storePath string, jsonOut bool, w io.Writer) error {
	st, info, err := openHistoryStore(storePath)
	if err != nil {
		return err
	}
	defer st.Close()

	failures, err := st.ListFailureReports(ctx, 0)
	if err != nil {
		return err
	}
	if failures == nil {
		failures = []store.FailureReport{}
	}
	if jsonOut {
		return writeJSON(w, reportJSON{
			Store:    historyStorePayload(info),
			Failures: failures,
		})
	}
	_, err = fmt.Fprint(w, formatReportText(failures))
	return err
}

func formatReportText(failures []store.FailureReport) string {
	if len(failures) == 0 {
		return "No stored failures yet.\nRun `faultline analyze <logfile>` a few times, then `faultline report` will show recurring failure classes.\n"
	}

	const maxEvidenceWidth = 72
	headers := []string{"Failure ID", "Count", "Last seen", "Example evidence"}
	widths := []int{len(headers[0]), len(headers[1]), len(headers[2]), len(headers[3])}
	rows := make([][]string, 0, len(failures))
	for _, item := range failures {
		row := []string{
			item.FailureID,
			fmt.Sprintf("%d", item.Count),
			item.LastSeenAt,
			truncateReportCell(item.ExampleEvidence, maxEvidenceWidth),
		}
		rows = append(rows, row)
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var b strings.Builder
	b.WriteString("Faultline Report\n\n")
	writeReportRow(&b, headers, widths)
	writeReportSeparator(&b, widths)
	for _, row := range rows {
		writeReportRow(&b, row, widths)
	}
	return b.String()
}

func writeReportRow(b *strings.Builder, cells []string, widths []int) {
	for i, cell := range cells {
		if i > 0 {
			b.WriteString("  ")
		}
		if i == 1 {
			fmt.Fprintf(b, "%*s", widths[i], cell)
			continue
		}
		fmt.Fprintf(b, "%-*s", widths[i], cell)
	}
	b.WriteByte('\n')
}

func writeReportSeparator(b *strings.Builder, widths []int) {
	for i, width := range widths {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(strings.Repeat("-", width))
	}
	b.WriteByte('\n')
}

func truncateReportCell(value string, limit int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if limit <= 0 || len(runes) <= limit {
		return value
	}
	if limit <= 3 {
		return string(runes[:limit])
	}
	return strings.TrimSpace(string(runes[:limit-3])) + "..."
}
