package main

import (
	"errors"
	"fmt"
	"os"

	"faultline/internal/app"
)

// Exit codes returned by the faultline CLI:
//
//	0 — success / no findings (guard: clean; batch: all logs matched)
//	1 — operational finding (guard: findings emitted; batch: one or more unmatched; analyze: silent failure with --fail-on-silent)
//	2 — error (invalid arguments, unreadable input, processing failure)
const (
	exitSuccess = 0
	exitFinding = 1
	exitError   = 2
)

func main() {
	if err := newRootCommand().Execute(); err != nil {
		switch {
		case errors.Is(err, app.ErrGuardFindings),
			errors.Is(err, app.ErrSilentFailure),
			errors.Is(err, app.ErrBatchUnmatched):
			// Operational result: output already written; signal exit 1.
			os.Exit(exitFinding)
		default:
			fmt.Fprintln(os.Stderr, err)
			os.Exit(exitError)
		}
	}
	os.Exit(exitSuccess)
}
