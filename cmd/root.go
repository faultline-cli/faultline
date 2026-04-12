package main

import (
	"github.com/spf13/cobra"

	"faultline/internal/cli"
)

// version is stamped at build time: -ldflags "-X main.version=x.y.z".
var version = "dev"

func newRootCommand() *cobra.Command {
	return cli.NewRootCommand(version)
}
