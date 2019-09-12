package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/davidovich/summon/internal/scaffold"
	"github.com/davidovich/summon/pkg/command"
	"github.com/spf13/cobra"
)

func main() {
	os.Exit(execute(newMainCmd()))
}

func execute(rootCmd *cobra.Command) int {
	err := rootCmd.Execute()

	if err != nil {
		return 1
	}
	return 0
}

var execCmd = command.New

func newMainCmd() *cobra.Command {
	var dest string
	var force bool
	var summonName string

	rootCmd := &cobra.Command{
		Use:   "scaffold",
		Short: "initialize an asset directory managed by summon",
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "scaffold [root go module name]",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := scaffold.Create(dest, args[0], summonName, force)
			if err == nil {
				fmt.Println("Successfully scaffolded a summon asset repo")
				git, err := exec.LookPath("git")
				if err != nil {
					fmt.Println("Warn: could not find git on PATH to initialize repository")
					return nil
				}
				gitcmd := execCmd(git, "-C", dest, "init")
				gitcmd.Stdout = os.Stdout
				_ = gitcmd.Run()
			}
			return err
		},
	}

	initCmd.Flags().StringVarP(&dest, "out", "o", ".", "destination directory")
	initCmd.Flags().StringVarP(&summonName, "name", "n", "summon", "summon executable name")
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "force overwrite")

	rootCmd.AddCommand(initCmd)
	return rootCmd
}
