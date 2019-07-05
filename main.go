package summon

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/davidovich/summon/cmd"
	"github.com/davidovich/summon/pkg/summon"
	"github.com/gobuffalo/packr/v2"
)

// Main entrypoint, typically called from a data repository
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
