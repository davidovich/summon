/*
Package summon is a library that allows giving a convenient command line to packed assets.

The binary created from this library is meant to be shared in a team to allow distribution
of common assets or code templates.

It solves the maintenance problem of multiple copies of same
code snippets distributed in many repos (like general makefile recipes), leveraging go modules and version
management. It also allows configuring a standard set of tools that a dev team can readily
invoke by name.

# Basics

This library needs a command entrypoint in a data repository. See https://github.com/davidovich/summon-example-assets.
It can be bootstrapped in an empty directory by using:

	cd [empty data repo dir]
	go run github.com/davidovich/summon/scaffold@latest init [module name]
*/
package summon

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/davidovich/summon/cmd"
	"github.com/davidovich/summon/pkg/summon"
)

// Main entrypoint, typically called from a data repository. Calling Main() relinquishes
// control to Summon so it can manage the command line arguments and instantiation of assets
// located in the embed.fs data repository parameter.
// Config opts functions are optional.
func Main(args []string, fs embed.FS, opts ...option) int {
	options := &MainOptions{}

	for _, o := range opts {
		o(options)
	}

	summon.Name = args[0]
	s, err := summon.New(fs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create initial filesystem: %v\n", err)
		return 1
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range c {
			os.Exit(0)
		}
	}()

	rootCmd, err := cmd.CreateRootCmd(s, os.Args, *options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create command tree: %v\n", err)
		return 1
	}

	err = rootCmd.Execute()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		}
		return 1
	}

	return 0
}

type option func(o *MainOptions)

// WithoutRunCmd configures summon to attach invocables directly to the
// main program. The default is to have these attached to the `run` subcommand.
func WithoutRunCmd() option {
	return func(o *MainOptions) {
		o.WithoutRunSubcmd = true
	}
}

// MainOptions hold compile-time configurations.
type MainOptions = summon.MainOptions
