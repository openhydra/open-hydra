package main

import (
	"os"

	"open-hydra/cmd/open-hydra-server/app"
)

var version string

func main() {
	// add a comment for ci test
	cmd := app.NewCommand(version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
