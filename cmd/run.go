package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/davidovich/summon/pkg/config"
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

func newRunCmd(runCmdDisabled bool, root *cobra.Command, driver summon.ConfigurableRunner, main *mainCmd) *cobra.Command {
	runCmd := &runCmdOpts{
		mainCmd: main,
		driver:  driver,
	}

	osArgs := os.Args
	if main.osArgs != nil {
		osArgs = *main.osArgs
	}

	// read config for exec section
	driver.Configure()
	invocables := []string{}

	handles := driver.ListInvocables()

	for h := range handles {
		invocables = append(invocables, h)
	}

	rcmd := root

	if !runCmdDisabled {
		rcmd = &cobra.Command{
			Use:   "run",
			Short: "Launch executable from summonables",
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				return invocables, cobra.ShellCompDirectiveNoFileComp
			},
			Args: func(cmd *cobra.Command, args []string) error {
				if len(args) < 1 {
					return fmt.Errorf("requires at least 1 command to run, received 0")
				}
				validArgs, _ := cmd.ValidArgsFunction(cmd, args, "")
				a := args[0]
				for _, v := range validArgs {
					if a == v {
						return nil
					}
				}
				return fmt.Errorf("invalid argument %q for %q", a, cmd.CommandPath())
			},
			FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
			Run:                func(cmd *cobra.Command, args []string) {},
		}
	}

	rcmd.PersistentFlags().BoolVarP(&runCmd.dryrun, "dry-run", "n", false, "only show what would be executed")

	makeRunCmd := func(summonRef string) func(cmd *cobra.Command, args []string) error {

		return func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			runCmd.ref = summonRef
			// calculate the extra args to pass to the referenced executable
			// this is due to a limitation in spf13/cobra which eats
			// all unknown args or flags making it hard to wrap other commands.
			// We are lucky, we know the prefix order of params,
			// extract args after the run command [summon run handle]
			// see https://github.com/spf13/pflag/pull/160
			// https://github.com/spf13/cobra/issues/739
			// and https://github.com/spf13/pflag/pull/199
			firstUnknownArgPos := 3
			if runCmdDisabled {
				firstUnknownArgPos = 2
			}
			runCmd.args = extractUnknownArgs(cmd.Flags(), osArgs[firstUnknownArgPos:])
			return runCmd.run()
		}
	}

	constructCommandTree(driver, rcmd, handles, makeRunCmd)

	if !runCmdDisabled && root != nil {
		root.AddCommand(rcmd)
	}
	return rcmd
}

func addCmdSpec(root *cobra.Command, arg string, cmdSpec config.CmdSpec, run func(cmd *cobra.Command, args []string) error) {
	subCmd := &cobra.Command{
		Use:                arg,
		RunE:               run,
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
	if cmdSpec.Args != nil {
		for aName, cmdSpec := range cmdSpec.Args {
			addCmdSpec(subCmd, aName, cmdSpec, run)
		}
	}
	if cmdSpec.Completion != "" {
		subCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			//cmd, err := driver.BuildCommand()
			//driver.Run(summon.Out(w io.Writer))
			return nil, cobra.ShellCompDirectiveDefault
		}
	}
	if len(cmdSpec.CmdArgs) != 0 {
		_ = 0
		// render template and FlattenResult
		// apply result to driver Args
	}
	if len(cmdSpec.Flags) != 0 {
		_ = 0
		// declare a storage for flags
		// add flags to cobra command
		// pass flags storage to Driver
	}

	subCmd.Short = cmdSpec.Help

	root.AddCommand(subCmd)
}

func constructCommandTree(driver summon.ConfigurableRunner, root *cobra.Command, handles config.Handles, makerun func(ref string) func(cmd *cobra.Command, args []string) error) {
	for h, args := range handles {
		switch t := args.Value.(type) {
		case config.CmdSpec:
			addCmdSpec(root, h, t, makerun(h))

		case config.ArgSliceSpec:
			subCmd := &cobra.Command{
				Use:                h,
				RunE:               makerun(h),
				FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
			}
			root.AddCommand(subCmd)
		}
	}
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
