/*
Package summon is a library that allows giving a convenient command line to packed assets.

The binary created from this library is meant to be shared in a team to allow distribution
of common assets or code templates.

It solves the maintenance problem of multiple copies of same
code snippets distributed in many repos (like general makefile recipies), leveraging go modules and version
management. It also allows configuring a standard set of tools that a dev team can readily
invoke by name.

Basics

This library needs a command entrypoint in a data repository. See https://github.com/davidovich/summon-example-assets.
It can be bootstrapped in an empty directory by using:

  cd [empty data repo dir]
  go run github.com/davidovich/summon/scaffold init [module name]
*/
package summon

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/davidovich/summon/cmd"
	"github.com/davidovich/summon/pkg/summon"
	"github.com/gobuffalo/packr/v2"
)

// Main entrypoint, typically called from a data repository. Calling Main() relinquishes
// control to Summon so it can manage the command line arguments and instanciation of assets
// located in the passed packr.Box data repository.
// See https://github.com/gobuffalo/packr/tree/master/v2 for more information on packr Boxes.
func Main(args []string, box *packr.Box) int {
	summon.Name = args[0]
	s, err := summon.New(box)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create initial box: %v", err)
		return 1
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range c {
			os.Exit(0)
		}
	}()

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
