package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/davidovich/summon/pkg/summon"
)

func newRunCmd(runCmdEnabled bool, root *cobra.Command, driver summon.ConfigurableRunner, main *mainCmd) (*cobra.Command, error) {
	osArgs := os.Args
	if main.osArgs != nil {
		osArgs = *main.osArgs
	}

	// read config for exec section
	driver.Configure(
		summon.Args(osArgs...),
	)

	return driver.ConstructCommandTree(root, runCmdEnabled)
}
