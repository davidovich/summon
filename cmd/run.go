package cmd

import (
	"fmt"
	"os"

	"github.com/davidovich/summon/pkg/summon"
	"github.com/spf13/cobra"
)

type runCmdOpts struct {
	driver summon.ConfigurableRunner
	ref    string
	args   []string
}

func newRunCmd(driver summon.ConfigurableRunner) *cobra.Command {
	runCmd := &runCmdOpts{
		driver: driver,
	}

	invocables := driver.ListInvocables()
	rcmd := &cobra.Command{
		Use:       "run",
		Short:     "Launch executable from summonables",
		ValidArgs: invocables,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("requires at least 1 command to run, received 0")
			}
			return cobra.ExactValidArgs(1)(cmd, args)
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		Run:                func(cmd *cobra.Command, args []string) {},
	}

	subRunE := func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		runCmd.ref = cmd.Name()
		// pass all Args down to the referenced executable
		// this is due to a limitation in spf13/cobra which eats
		// all unknown args or flags making it hard to wrap other commands.
		// We are lucky, we know the structure, just pass all args.
		// see https://github.com/spf13/pflag/pull/160
		runCmd.args = os.Args[3:] // 3 is [summon, run, handle]
		return runCmd.run()
	}
	for _, i := range invocables {
		runSubCmd := &cobra.Command{
			Use:                i,
			RunE:               subRunE,
			FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		}
		rcmd.AddCommand(runSubCmd)
	}

	return rcmd
}

func (r *runCmdOpts) run() error {
	r.driver.Configure(
		summon.Ref(r.ref),
		summon.Args(r.args...),
	)

	return r.driver.Run()
}
