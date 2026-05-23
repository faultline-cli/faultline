package cli

import (
	"fmt"

	"faultline/internal/output"
)

func validateOutputFormat(value string) (output.Format, error) {
	format, ok := output.ParseFormat(value)
	if !ok {
		return "", fmt.Errorf("--format must be %q, %q, or %q", output.FormatTerminal, output.FormatMarkdown, output.FormatJSON)
	}
	return format, nil
}

func validateOutputMode(value string) error {
	if value != string(output.ModeQuick) && value != string(output.ModeDetailed) {
		return fmt.Errorf("--mode must be %q or %q", output.ModeQuick, output.ModeDetailed)
	}
	return nil
}

func validateSelect(value int) error {
	if value < 0 {
		return fmt.Errorf("--select must be 1 or greater")
	}
	return nil
}

func resolveOutputSelection(formatValue string, jsonOut bool) (output.Format, bool, error) {
	format, err := validateOutputFormat(formatValue)
	if err != nil {
		return "", false, err
	}
	if jsonOut {
		if format != output.FormatTerminal && format != output.FormatJSON {
			return "", false, fmt.Errorf("--json cannot be combined with --format %q", format)
		}
		format = output.FormatJSON
	}
	return format, format == output.FormatJSON, nil
}
