package service

import (
	"fmt"
	"os/exec"
)

func SyncRemote() error {
	cmd := exec.Command("git", "fetch", "--all")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fetch refs: %w", err)
	}
	return nil
}
