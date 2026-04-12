package main

import (
	"fmt"
	"os"

	"faultline/internal/playbooks"
)

func main() {
	pbs, err := playbooks.NewCatalog("").Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	conflicts := playbooks.FindPatternConflicts(pbs)
	fmt.Print(playbooks.FormatPatternConflicts(conflicts))
}
