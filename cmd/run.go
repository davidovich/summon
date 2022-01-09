package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/shlex"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/davidovich/summon/pkg/config"
	"github.com/davidovich/summon/pkg/summon"
)

type runCmdOpts struct {
	*mainCmd
	driver   summon.ConfigurableRunner
	ref      string
	args     []string
	dryrun   bool
	userArgs []string
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
	if firstUnknownArgPos >= len(osArgs) {
		firstUnknownArgPos = len(osArgs)
	}
	runCmd.userArgs = osArgs[firstUnknownArgPos:]

	// read config for exec section
	driver.Configure()
	flags, handles := driver.ExecContext()

	rootcmd := root

	if !runCmdDisabled {
		rootcmd = &cobra.Command{
			Use:   "run [handle]",
			Short: "Launch executable from summonables",
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				invocables := make([]string, 0, len(handles))
				for h := range handles {
					invocables = append(invocables, h)
				}
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

	for flagName, flag := range flags {
		_ = flagName
		_ = flag
	}

	rootcmd.PersistentFlags().BoolVarP(&runCmd.dryrun, "dry-run", "n", false, "only show what would be executed")

	runCmd.constructCommandTree(rootcmd, handles)

	if !runCmdDisabled && root != nil {
		root.AddCommand(rootcmd)
	}
	return rootcmd
}

func (r *runCmdOpts) addCmdSpec(root *cobra.Command, arg string, cmdSpec config.CmdSpec, run func(*cobra.Command, []string) error) {
	subCmd := &cobra.Command{
		Use:                arg,
		RunE:               run,
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
	if cmdSpec.Args != nil {
		for cName, cmdSpec := range cmdSpec.Args {
			r.addCmdSpec(subCmd, cName, cmdSpec, run)
		}
	}
	if cmdSpec.Completion != "" {
		subCmd.ValidArgsFunction = func(cmd *cobra.Command, cobraArgs []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			r.driver.Configure(summon.Args(extractUnknownArgs(cmd.Flags(), r.userArgs)...))
			args, err := r.driver.RenderArgs(cmdSpec.Completion)
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return nil, cobra.ShellCompDirectiveError
			}
			splitArgs, err := shlex.Split(strings.Join(args, " "))
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			candidates := []string{}
			for _, a := range splitArgs {
				if strings.Contains(a, toComplete) {
					candidates = append(candidates, a)
				}
			}
			return candidates, cobra.ShellCompDirectiveDefault
		}
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

func (r *runCmdOpts) constructCommandTree(root *cobra.Command, handles config.Handles) {
	makerun := func(summonRef string) func(cmd *cobra.Command, args []string) error {
		runCmd := runCmdOpts{
			mainCmd:  r.mainCmd,
			driver:   r.driver,
			ref:      summonRef,
			userArgs: r.userArgs,
		}
		return func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			runCmd.dryrun = r.dryrun
			runCmd.args = extractUnknownArgs(cmd.Flags(), runCmd.userArgs)
			return runCmd.run()
		}
	}

	for h, args := range handles {
		switch t := args.Value.(type) {
		case config.CmdSpec:
			r.addCmdSpec(root, h, t, makerun(h))

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
		if len(a) > 0 && a[0] == '-' && len(a) > 1 {
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
