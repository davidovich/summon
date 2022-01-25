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

// commandSpec describes a normalized command
type commandSpec struct {
	// execEnvironment is the caller environment (docker, bash, python)
	execEnvironment string
	// args is the command and args that get appended to the ExecEnvironment
	args config.ArgSliceSpec
	// subCmd sub-command of current command
	subCmd map[string]*commandSpec
	// flags of this command
	flags config.Flags
	// help of this command
	help string
	// Command to invoke to have a completion of this command
	completion string
	// hidden hides the command from help
	hidden bool
}

// handles are the normalized version of the configs HandleDesc
type handles map[string]*commandSpec

// Run will run executable scripts described in the summon.config.yaml file
// of the data repository module.
func (d *Driver) Run(opts ...Option) error {
	err := d.Configure(opts...)
	if err != nil {
		return err
	}

	cmdArgs, err := d.buildCmdArgs()
	if err != nil {
		return err
	}

	cmd := d.execCommand(cmdArgs[0], cmdArgs[1:]...)
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

		return cmd.Run()
	}

	return nil
}

func (d *Driver) buildCmdArgs() ([]string, error) {
	var cmdSpec *commandSpec
	var ref string
	if d.opts.cobraCmd != nil {
		cmdSpec = d.cmdToSpec[d.opts.cobraCmd]
		ref = d.opts.cobraCmd.Name()
	} else {
		// find the corresponding command
		for cmd, spec := range d.cmdToSpec {
			if cmd.Name() == d.opts.ref {
				cmdSpec = spec
				ref = d.opts.ref
				break
			}
		}
	}
	if cmdSpec == nil {
		return nil, fmt.Errorf("could not find exec handle reference '%s' in config %s", ref, config.ConfigFileName)
	}

	renderedExecEnv, err := d.renderTemplate(cmdSpec.execEnvironment)
	if err != nil {
		return nil, err
	}
	execEnv, err := shlex.Split(renderedExecEnv)
	if err != nil {
		return nil, err
	}
	// Render and flatten arguments array of arrays to simple array
	arguments, err := d.RenderArgs(cmdSpec.args...)
	if err != nil {
		return nil, err
	}
	finalCmd := append(execEnv, arguments...)

	// Render flags
	renderedFlags := []string{}
	for _, flag := range d.flagsToRender {
		// if the flag was used in a template call do not use it implicitely
		if flag.explicit {
			continue
		}
		renderedFlag, err := flag.renderTemplate()
		if err != nil {
			return nil, err
		}
		renderedFlags = append(renderedFlags, renderedFlag)
	}

	finalCmd = append(finalCmd, renderedFlags...)
	// add user args that were not consumed by a template render
	unusedArgs := computeUnused(d.opts.args, d.opts.argsConsumed)
	finalCmd = append(finalCmd, unusedArgs...)

	return finalCmd, nil
}

func (d *Driver) RenderArgs(args ...interface{}) ([]string, error) {
	targets := make([]string, 0, len(args))
	for _, t := range FlattenStrings(args) {
		rt, err := d.renderTemplate(t)
		if err != nil {
			return nil, err
		}
		if rt == "" {
			continue
		}

		renderedTargets := []string{rt}
		if strings.HasPrefix(rt, "[") && strings.HasSuffix(rt, "]") {
			inner := strings.Trim(rt, "[]")
			renderedTargets = strings.Split(inner, "\n")
			if inner == "" {
				renderedTargets = []string{""}
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

func normalizeExecDesc(argsDesc interface{}, invoker string) (*commandSpec, error) {
	c := &commandSpec{}
	switch descType := argsDesc.(type) {
	case config.ArgSliceSpec:
		c.args = descType
	case config.CmdDesc:
		c.args = descType.Args
		c.help = descType.Help
		c.completion = descType.Completion
		c.hidden = descType.Hidden
		if descType.SubCmd != nil {
			c.subCmd = make(map[string]*commandSpec)
			for subCmdName, execDesc := range descType.SubCmd {
				subCmd, err := normalizeExecDesc(execDesc.Value, invoker)
				if err != nil {
					return nil, err
				}
				c.subCmd[subCmdName] = subCmd
			}
		}
		c.flags = normalizeFlags(descType.Flags)
	default:
		return nil, fmt.Errorf("in config %s: unhandled type: %T",
			config.ConfigFileName, descType)
	}

	c.execEnvironment = invoker
	return c, nil
}

// execContext lists the execEnvironments in the config file under the exec:
// key.
func (d *Driver) execContext() (config.Flags, handles, error) {
	if d.globalFlags == nil {
		d.globalFlags = normalizeFlags(d.config.Exec.GlobalFlags)
	}

	if d.handles == nil {
		handles := handles{}
		for invoker, handleDescs := range d.config.Exec.ExecEnv {
			for handle, desc := range handleDescs {
				if _, present := handles[handle]; present {
					return nil, nil,
						fmt.Errorf("config error for 'exec.environments:%s' in config %s: cannot have duplicate handles: '%s'", invoker, config.ConfigFileName, handle)
				}
				cmdSpec, err := normalizeExecDesc(desc.Value, invoker)
				if err != nil {
					return nil, nil, fmt.Errorf("error in exec:environments:%s %s", invoker, err.Error())
				}
				handles[handle] = cmdSpec
			}
		}
		d.handles = handles
	}

	return d.globalFlags, d.handles, nil
}

func normalizeFlags(flagsDesc map[string]config.FlagDesc) config.Flags {
	normalizedFlags := config.Flags{}
	for flagName, flags := range flagsDesc {
		switch f := flags.Value.(type) {
		case string:
			normalizedFlags[flagName] = &config.FlagSpec{
				Effect: f,
			}
		case config.FlagSpec:
			normalizedFlags[flagName] = &f
		}
	}
	return normalizedFlags
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

const (
	global bool = true
	local  bool = false
)

func (d *Driver) ConstructCommandTree(root *cobra.Command, runCmdEnabled bool) error {
	globalFlags, handles, err := d.execContext()
	if err != nil {
		return err
	}

	if runCmdEnabled {
		newRoot := &cobra.Command{
			Use:                "run [handle]",
			Short:              "Launch executable from summonables",
			FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
			Run:                func(cmd *cobra.Command, args []string) {},
			Args: func(cmd *cobra.Command, args []string) error {
				if len(args) < 1 {
					return fmt.Errorf("requires at least 1 command to run, received 0")
				}
				a := args[0]
				if _, ok := handles[a]; !ok {
					return fmt.Errorf("invalid argument %q for %q", a, cmd.CommandPath())
				}
				return nil
			},
		}

		root.AddCommand(newRoot)
		root = newRoot
	}
	root.PersistentFlags().BoolVarP(&d.opts.dryrun, "dry-run", "n", false, "only show what would be executed")

	d.AddFlags(root, globalFlags, global)

	for h, spec := range handles {
		d.addCmdSpec(root, h, spec)
	}

	d.setupArgs(root, runCmdEnabled)

	return nil
}

func (d *Driver) addCmdSpec(root *cobra.Command, arg string, cmdSpec *commandSpec) {
	subCmd := &cobra.Command{
		Use: arg,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			return d.Run(CobraCmd(cmd),
				Args(extractUnknownArgs(cmd.Flags(), d.opts.args)...))
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
	if cmdSpec.subCmd != nil {
		for cName, cmdSpec := range cmdSpec.subCmd {
			d.addCmdSpec(subCmd, cName, cmdSpec)
		}
	}
	if cmdSpec.completion != "" {
		subCmd.ValidArgsFunction = func(cmd *cobra.Command, cobraArgs []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			d.Configure(Args(extractUnknownArgs(cmd.Flags(), d.opts.args)...))
			inlineComp, err := d.RenderArgs(cmdSpec.completion)
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return nil, cobra.ShellCompDirectiveError
			}

			var completions, candidates []string
			for _, comp := range inlineComp {
				comp = strings.TrimRight(comp, "\n")
				candidates = append(candidates, strings.Split(comp, "\n")...)
			}

			for _, candidate := range candidates {
				// check that this argument was not completed before, if it
				// is, it will appear as a cobraArg of this command, and should
				// not be repeated (neither others in the candidates)
				for _, userArg := range cobraArgs {
					if userArg == candidate {
						return nil, cobra.ShellCompDirectiveDefault
					}
				}
				if strings.HasPrefix(candidate, toComplete) {
					completions = append(completions, candidate)
				}
			}
			return completions, cobra.ShellCompDirectiveDefault
		}
	}

	d.AddFlags(subCmd, cmdSpec.flags, local)
	d.cmdToSpec[subCmd] = cmdSpec

	subCmd.Short = cmdSpec.help
	subCmd.Hidden = cmdSpec.hidden

	root.AddCommand(subCmd)
}

func (d *Driver) setupArgs(root *cobra.Command, withRunCmd bool) {
	// all args after arg[0] which is the main program name
	if len(d.opts.args) == 0 {
		panic("missing Args call to Configure")
	}
	allArgs := d.opts.args[1:]

	root.ParseFlags(allArgs)
	root.Root().PersistentPreRun = func(cmd *cobra.Command, args []string) {
		_, d.opts.args, _ = cmd.Root().Find(allArgs)
	}

	root.Root().SetArgs(allArgs)
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
