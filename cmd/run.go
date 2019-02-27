package cmd

import (
	"github.com/davidovich/summon/pkg/summon"
	"github.com/spf13/cobra"
)

type runCmdOpts struct {
	driver summon.Interface
	ref    string
}

func newRunCmd(driver summon.Interface) *cobra.Command {
	runCmd := &runCmdOpts{
		driver: driver,
	}
	rcmd := &cobra.Command{
		Use:   "run",
		Short: "launch executable from summonables",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runCmd.ref = args[0]
			cmd.SilenceUsage = true
			return runCmd.run()
		},
	}

	return rcmd
}

func (r *runCmdOpts) run() error {
	r.driver.Configure(summon.Ref(r.ref))

	return r.driver.Run()
}
