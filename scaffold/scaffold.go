package main

import (
	"os"
	"os/exec"

	"github.com/davidovich/summon/pkg/scaffold"
)

func main() {
	err := scaffold.Create(".", false)

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
	}
	os.Exit(0)
}
