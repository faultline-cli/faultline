//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func main() {
	var files []string
	for _, pat := range []string{
		"playbooks/auth/*.yaml",
		"playbooks/build/*.yaml",
		"playbooks/test/*.yaml",
		"playbooks/network/*.yaml",
		"playbooks/runtime/*.yaml",
		"playbooks/deploy/*.yaml",
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
}
