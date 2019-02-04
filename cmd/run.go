package cmd

import (
	"github.com/davidovich/summon/pkg/summon"
	"github.com/gobuffalo/packr/v2"
	"github.com/spf13/cobra"
)

type runCmdOpts struct {
	box    *packr.Box
	driver summon.Interface
}

func newRunCmd(box *packr.Box, driver summon.Interface) *cobra.Command {
	runCmd := &runCmdOpts{
		box:    box,
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

func (l *runCmdOpts) run() error {
	if l.driver == nil {
		l.driver = summon.New(l.box)
	}

	return l.driver.Run()
}
