package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/davidovich/summon/pkg/summon"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type runCmdOpts struct {
	*mainCmd
	driver summon.ConfigurableRunner
	ref    string
	args   []string
	dryrun bool
}

func newRunCmd(root *cobra.Command, driver summon.ConfigurableRunner, main *mainCmd) *cobra.Command {
	runCmd := &runCmdOpts{
		mainCmd: main,
		driver:  driver,
	}

	osArgs := os.Args
	if main.osArgs != nil {
		osArgs = *main.osArgs
	}

	driver.Configure()

	invocables := driver.ListInvocables()

	rcmd := root

	if !driver.RunCmdDisabled() {
		rcmd = &cobra.Command{
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
	}

	rcmd.PersistentFlags().BoolVarP(&runCmd.dryrun, "dry-run", "n", false, "only show what would be executed")

	firstUnknownArgPos := 3
	if driver.RunCmdDisabled() {
		firstUnknownArgPos = 2
	}
	subRunE := func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		runCmd.ref = cmd.Name()
		// calculate the extra args to pass to the referenced executable
		// this is due to a limitation in spf13/cobra which eats
		// all unknown args or flags making it hard to wrap other commands.
		// We are lucky, we know the prefix order of params,
		// extract args after the run command [summon run handle]
		// see https://github.com/spf13/pflag/pull/160
		// https://github.com/spf13/cobra/issues/739
		// and https://github.com/spf13/pflag/pull/199
		runCmd.args = extractUnknownArgs(cmd.Flags(), osArgs[firstUnknownArgPos:])
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

	if !driver.RunCmdDisabled() && root != nil {
		root.AddCommand(rcmd)
	}
	return rcmd
}

func extractUnknownArgs(flags *pflag.FlagSet, args []string) []string {
	unknownArgs := []string{}

	for i := 0; i < len(args); i++ {
		a := args[i]
		var f *pflag.Flag
		if a[0] == '-' && len(a) > 1 {
			if a[1] == '-' {
				f = flags.Lookup(strings.SplitN(a[2:], "=", 2)[0])
			} else {
				for _, s := range a[1:] {
					f = flags.ShorthandLookup(string(s))
					if f == nil {
						break
					}
				}
			}
		}
		if f != nil {
			if f.NoOptDefVal == "" && i+1 < len(args) && f.Value.String() == args[i+1] {
				i++
			}
			continue
		}
		unknownArgs = append(unknownArgs, a)
	}
	return unknownArgs
}

func (r *runCmdOpts) run() error {
	err := r.driver.Configure(
		summon.Ref(r.ref),
		summon.Args(r.args...),
		summon.JSON(r.json),
		summon.Debug(r.debug),
		summon.DryRun(r.dryrun),
	)

	if err != nil {
		return err
	}

	return r.driver.Run()
}
