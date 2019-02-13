package cmd

import (
	"github.com/davidovich/summon/pkg/summon"
	"github.com/spf13/cobra"
)

type runCmdOpts struct {
	driver summon.Interface
}

func newRunCmd(driver summon.Interface) *cobra.Command {
	runCmd := &runCmdOpts{
		driver: driver,
	}
	rcmd := &cobra.Command{
		Use:   "run",
		Short: "launch executable from summonables",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmd.run()
		},
	}

	return rcmd
}

func (r *runCmdOpts) run() error {
	r.driver.Configure()

	return r.driver.Run()
}
