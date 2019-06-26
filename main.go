/*
Package summon is a library allowing to pack assets and give it a convienient command line.

The binary created from this library is meant to be shared in a team to allow distribution
of common assets or code templates.

Basics

This library needs a command entrypoint in a data repository. See https://github.com/davidovich/summon-example-assets.
*/
package summon

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/davidovich/summon/cmd"
	"github.com/davidovich/summon/pkg/summon"
	"github.com/gobuffalo/packr/v2"
)

// Main entrypoint, typically called from a data repository. Calling Main() relinquishes
// control to Summon so it can manage the command line arguments and instanciation of assets in
// the packr.Box data repository.
func Main(args []string, box *packr.Box) int {
	s, err := summon.New(box)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create initial box: %v", err)
		return 1
	}

	rootCmd := cmd.CreateRootCmd(s, os.Args)
	err = rootCmd.Execute()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		}
		return 1
	}

	return 0
}
