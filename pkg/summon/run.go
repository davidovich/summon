package summon

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/google/shlex"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/davidovich/summon/pkg/config"
)

type execUnit struct {
	invoker     string
	invokerArgs string
	env         []string
	flags       *config.FlagSpec
	targetSpec  *config.CmdSpec
}

// Run will run executable scripts described in the summon.config.yaml file
// of the data repository module.
func (d *Driver) Run(opts ...Option) error {
	err := d.Configure(opts...)
	if err != nil {
		return err
	}

	env, rargs, err := d.BuildCommand()
	if err != nil {
		return err
	}

	cmd := d.execCommand(rargs[0], rargs[1:]...)

	if d.opts.debug || d.opts.dryrun {
		msg := "Executing"
		if d.opts.dryrun {
			msg = "Would execute"
		}
		fmt.Fprintf(os.Stderr, "%s `%s`...\n", msg, cmd)
	}

	if !d.opts.dryrun {
		cmd.Stdin = os.Stdin
		cmd.Stdout = d.opts.out
		cmd.Stderr = os.Stderr
		cmd.Env = append(cmd.Env, env...)

		return cmd.Run()
	}

	return nil
}

func (d *Driver) BuildCommand() ([]string, []string, error) {
	eu, err := d.findExecutor(d.opts.ref)
	if err != nil {
		return nil, nil, err
	}

	invArgs, err := d.renderTemplate(eu.invokerArgs)
	if err != nil {
		return nil, nil, err
	}

	// See if we have an overridden command in the config.
	// Each user-supplied args is tried in order to see if we have
	// an override. If there is an override, this arg is consumed, Otherwize
	// it is kept for the downstream commmand construction.
	cmdSpec := *eu.targetSpec
	if cmdSpec.Args != nil {
		newArgs := []string{}
		for _, a := range d.opts.args {
			newCmdSpec, ok := cmdSpec.Args[a]
			if ok {
				// we have an override for this arg, try going deeper
				cmdSpec = newCmdSpec
			} else {
				// no override, keep the user provided arg
				newArgs = append(newArgs, a)
			}
		}
		d.opts.args = newArgs
	}

	// Render and flatten arguments array of arrays to simple array
	targets, err := d.RenderArgs(cmdSpec.CmdArgs...)
	if err != nil {
		return nil, nil, err
	}

	rargs := []string{eu.invoker}
	opts, err := shlex.Split(invArgs)
	if err != nil {
		return nil, nil, err
	}
	rargs = append(rargs, opts...)
	rargs = append(rargs, targets...)

	unusedArgs := computeUnused(d.opts.args, d.opts.argsConsumed)
	rargs = append(rargs, unusedArgs...)

	return eu.env, rargs, nil
}

func (d *Driver) RenderArgs(args ...interface{}) ([]string, error) {
	targets := make([]string, 0, len(args))
	var renderedTargets []string
	for _, t := range FlattenStrings(args) {
		rt, err := d.renderTemplate(t)
		if err != nil {
			return nil, err
		}

		renderedTargets = []string{rt}

		if strings.HasPrefix(rt, "[") && strings.HasSuffix(rt, "]") {
			renderedTargets, err = shlex.Split(strings.Trim(rt, "[]"))
			if err != nil {
				return nil, err
			}
		}

		targets = append(targets, renderedTargets...)
	}
	return targets, nil
}

func computeUnused(args []string, consumed map[int]struct{}) []string {
	unusedArgs := []string{}
	if len(consumed) == len(args) {
		return unusedArgs
	}
	for i, a := range args {
		if _, ok := consumed[i]; ok {
			continue
		}
		unusedArgs = append(unusedArgs, a)
	}
	return unusedArgs
}

// ExecContext lists the invokers in the config file under the exec:
// key.
func (d *Driver) ExecContext() (map[string]config.FlagSpec, config.Handles) {
	handles := config.Handles{}

	normalizedFlags := normalizeFlags(d.Config.Exec.GlobalFlags)
	for _, invokers := range d.Config.Exec.Invokers {
		for i, v := range invokers {
			handles[i] = v
		}
	}

	return normalizedFlags, handles
}

func normalizeFlags(flagsDesc map[string]config.FlagDesc) map[string]config.FlagSpec {
	normalizedFlags := map[string]config.FlagSpec{}
	for flagName, flags := range flagsDesc {
		switch f := flags.Value.(type) {
		case string:
			normalizedFlags[flagName] = config.FlagSpec{
				Effect: f,
			}
		case config.FlagSpec:
			normalizedFlags[flagName] = f
		}
	}
	return normalizedFlags
}

func (d *Driver) findExecutor(ref string) (execUnit, error) {
	eu := execUnit{}

	// Extract env part if present
	env, err := shlex.Split(ref)
	if err != nil {
		return execUnit{}, err
	}
	handleIndex := 0
	for i, e := range env {
		if !strings.Contains(e, "=") {
			handleIndex = i
			break
		}
		eu.env = append(eu.env, e)
	}
	handle := env[handleIndex]

	for invoker, handles := range d.Config.Exec.Invokers {
		if h, ok := handles[handle]; ok {
			if eu.invoker != "" {
				return execUnit{}, fmt.Errorf("config syntax error for 'exec.invokers:%s' in config %s: cannot have duplicate handles: '%s'", invoker, config.ConfigFile, handle)
			}
			exec := strings.SplitAfterN(invoker, " ", 2)
			eu.invoker = strings.TrimSpace(exec[0])
			if len(exec) == 2 {
				eu.invokerArgs = strings.TrimSpace(exec[1])
			}

			spec := config.CmdSpec{}
			switch s := h.Value.(type) {
			case config.ArgSliceSpec:
				spec.CmdArgs = s
			case config.CmdSpec:
				spec = s
			default:
				return execUnit{}, fmt.Errorf("config syntax error for 'exec.invokers:%s:%s' in config %s", invoker, handle, config.ConfigFile)
			}
			eu.targetSpec = &spec
		}
	}

	if eu.invoker == "" {
		return eu, fmt.Errorf("could not find exec handle reference '%s' in config %s", handle, config.ConfigFile)
	}

	return eu, nil
}

func flatten(args []interface{}, v reflect.Value) []interface{} {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	if v.Kind() == reflect.Array || v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			args = flatten(args, v.Index(i))
		}
	} else {
		args = append(args, v.Interface())
	}

	return args
}

// FlattenStrings takes an array of string values or string slices and returns
// an flattened slice of strings.
func FlattenStrings(args ...interface{}) []string {
	flattened := flatten(nil, reflect.ValueOf(args))
	s := make([]string, 0, len(flattened))
	for _, f := range flattened {
		s = append(s, f.(string))
	}
	return s
}

func (d *Driver) ConstructCommandTree(root *cobra.Command, runCmdDisabled bool) *cobra.Command {

	_, handles := d.ExecContext()

	if !runCmdDisabled {
		newRoot := &cobra.Command{
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

		if root != nil {
			root.AddCommand(newRoot)
		}
		root = newRoot
	}
	root.PersistentFlags().BoolVarP(&d.opts.dryrun, "dry-run", "n", false, "only show what would be executed")

	makerun := func(summonRef string) func(cmd *cobra.Command, args []string) error {
		return func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			return d.Run(Ref(summonRef),
				Args(extractUnknownArgs(cmd.Flags(), d.opts.args)...))
		}
	}

	for h, args := range handles {
		switch t := args.Value.(type) {
		case config.CmdSpec:
			d.addCmdSpec(root, h, t, makerun(h))

		case config.ArgSliceSpec:
			subCmd := &cobra.Command{
				Use:                h,
				RunE:               makerun(h),
				FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
			}
			root.AddCommand(subCmd)
		}
	}
	return root
}

func (d *Driver) addCmdSpec(root *cobra.Command, arg string, cmdSpec config.CmdSpec, run func(*cobra.Command, []string) error) {
	subCmd := &cobra.Command{
		Use:                arg,
		RunE:               run,
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
	if cmdSpec.Args != nil {
		for cName, cmdSpec := range cmdSpec.Args {
			d.addCmdSpec(subCmd, cName, cmdSpec, run)
		}
	}
	if cmdSpec.Completion != "" {
		subCmd.ValidArgsFunction = func(cmd *cobra.Command, cobraArgs []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			d.Configure(Args(extractUnknownArgs(cmd.Flags(), d.opts.args)...))
			args, err := d.RenderArgs(cmdSpec.Completion)
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
		//subCmd.PersistentFlags().String(name string, value string, usage string)
		// declare a storage for flags
		// add flags to cobra command
		// pass flags storage to Driver
	}

	subCmd.Short = cmdSpec.Help

	root.AddCommand(subCmd)
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
