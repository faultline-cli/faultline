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
	_ = filepath.Walk("playbooks/bundled", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			files = append(files, path)
		}
		return nil
	})
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

	pbs, err := playbooks.NewCatalog("playbooks/bundled").Load()
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
