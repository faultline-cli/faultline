//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"faultline/internal/playbooks"
)

func main() {
	var files []string
	for _, pat := range []string{
		"playbooks/auth/*.yaml",
		"playbooks/build/*.yaml",
		"playbooks/ci/*.yaml",
		"playbooks/deploy/*.yaml",
		"playbooks/network/*.yaml",
		"playbooks/runtime/*.yaml",
		"playbooks/test/*.yaml",
	} {
		matches, _ := filepath.Glob(pat)
		files = append(files, matches...)
	}
	ok := true
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			fmt.Printf("READ ERROR %s: %v\n", f, err)
			ok = false
			continue
		}
		var m map[string]interface{}
		if err := yaml.Unmarshal(data, &m); err != nil {
			fmt.Printf("YAML ERROR %s: %v\n", f, err)
			ok = false
		}
	}
	if ok {
		fmt.Printf("All %d playbooks parsed OK\n", len(files))
	}

	pbs, err := playbooks.LoadDir("playbooks")
	if err != nil {
		fmt.Printf("OVERLAP CHECK ERROR: %v\n", err)
		return
	}
	conflicts := playbooks.FindPatternConflicts(pbs)
	fmt.Printf("Found %d pattern conflicts\n", len(conflicts))
	if len(conflicts) > 0 {
		fmt.Print(playbooks.FormatPatternConflicts(conflicts))
	}
}
