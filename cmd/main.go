package main

import (
	"errors"
	"fmt"
	"os"

	"faultline/internal/app"
)

func main() {
	if err := newRootCommand().Execute(); err != nil {
		if !errors.Is(err, app.ErrGuardFindings) && !errors.Is(err, app.ErrSilentFailure) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}
