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
	cmdSpec, err := d.findExecutor(d.opts.ref)
	if err != nil {
		return nil, err
	}

	// See if we have an overridden command in the config.
	// Each user-supplied args is tried in order to see if we have
	// an override. If there is an override, this arg is consumed, Otherwize
	// it is kept for the downstream commmand construction.
	//
	// TODO: might be interesting to use the cobra command tree directly here.
	// Historically, there was only one entry point to select the command
	// configuration (the unique execution handle). Now there is a possibility
	// to describe command trees in the config and the user will invoke
	// a cobra based command string, which already has the correct parsing
	// and parenting from the command line.
	if cmdSpec.Args != nil {
		newArgs := []string{}
		for _, a := range d.opts.args {
			newCmdSpec, ok := cmdSpec.Args[a]
			if ok {
				// we have an override for this arg, try going deeper
				newCmdSpec.ExecEnvironment = cmdSpec.ExecEnvironment
				cmdSpec = newCmdSpec
			} else {
				// no override, keep the user provided arg
				newArgs = append(newArgs, a)
			}
		}
		d.opts.args = newArgs
	}

	renderedExecEnv, err := d.renderTemplate(cmdSpec.ExecEnvironment)
	if err != nil {
		return nil, err
	}
	execEnv, err := shlex.Split(renderedExecEnv)
	if err != nil {
		return nil, err
	}
	// Render and flatten arguments array of arrays to simple array
	arguments, err := d.RenderArgs(cmdSpec.Cmd...)
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
			renderedTargets, err = shlex.Split(inner)
			if err != nil {
				return nil, err
			}
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

// execContext lists the invokers in the config file under the exec:
// key.
func (d *Driver) execContext() (config.Flags, config.Handles, error) {
	if d.globalFlags == nil {
		d.globalFlags = normalizeFlags(d.config.Exec.GlobalFlags)
	}

	if d.handles == nil {
		handles := config.Handles{}
		for invoker, handleDescs := range d.config.Exec.Invokers {
			for handle, desc := range handleDescs {
				if _, present := handles[handle]; present {
					return config.Flags{}, config.Handles{},
						fmt.Errorf("config error for 'exec.invokers:%s' in config %s: cannot have duplicate handles: '%s'", invoker, config.ConfigFileName, handle)
				}
				switch descType := desc.Value.(type) {
				case config.ArgSliceSpec:
					c := &config.CmdSpec{}
					c.Cmd = descType
					c.ExecEnvironment = invoker
					handles[handle] = c
				case config.CmdSpec:
					descType.ExecEnvironment = invoker
					handles[handle] = &descType
				default:
					return config.Flags{}, config.Handles{},
						fmt.Errorf("config error for 'exec:invokers:%s in config %s: unhandled type: %T",
							invoker, config.ConfigFileName, descType)
				}
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

func (d *Driver) findExecutor(ref string) (*config.CmdSpec, error) {
	_, handles, err := d.execContext()
	if err != nil {
		return nil, err
	}

	if spec, ok := handles[ref]; ok {
		return spec, nil
	}

	return nil, fmt.Errorf("could not find exec handle reference '%s' in config %s", ref, config.ConfigFileName)
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

func (d *Driver) ConstructCommandTree(root *cobra.Command, runCmdEnabled bool) (*cobra.Command, error) {

	globalFlags, handles, err := d.execContext()
	if err != nil {
		return nil, err
	}

	if runCmdEnabled {
		newRoot := &cobra.Command{
			Use:   "run [handle]",
			Short: "Launch executable from summonables",
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
			FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
			Run:                func(cmd *cobra.Command, args []string) {},
		}

		if root != nil {
			root.AddCommand(newRoot)
		}
		root = newRoot
	}
	root.PersistentFlags().BoolVarP(&d.opts.dryrun, "dry-run", "n", false, "only show what would be executed")

	d.AddFlags(root, globalFlags)

	makerun := func(summonRef string) func(cmd *cobra.Command, args []string) error {
		return func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			return d.Run(Ref(summonRef),
				Args(extractUnknownArgs(cmd.Flags(), d.opts.args)...))
		}
	}

	for h, spec := range handles {
		d.addCmdSpec(root, h, spec, makerun(h))
	}
	return root, nil
}

func (d *Driver) AddFlags(cmd *cobra.Command, flags config.Flags) {
	for f, flagSpec := range flags {
		v := &flagValue{
			name:   f,
			d:      d,
			effect: flagSpec.Effect,
			// userValue: flagSpec.Default,
			explicit: flagSpec.Explicit,
		}
		flag := cmd.PersistentFlags().VarPF(v, f, flagSpec.Shorthand, flagSpec.Help)
		flag.NoOptDefVal = flagSpec.Default
	}
}

type flagValue struct {
	d         *Driver
	name      string
	effect    string
	userValue string
	rendered  string
	explicit  bool
}

func (f *flagValue) Set(s string) error {
	if f.d.flagsToRender == nil {
		f.d.flagsToRender = []*flagValue{}
	}
	f.d.flagsToRender = append(f.d.flagsToRender, f)
	f.userValue = s
	return nil
}

// String returns the current value
func (f *flagValue) String() string {
	return f.userValue
}

func (f *flagValue) Type() string {
	return "string"
}

func (f *flagValue) renderTemplate() (string, error) {
	if f.rendered != "" {
		return f.rendered, nil
	}
	var err error
	f.d.opts.data["flag"] = f.userValue
	f.rendered, err = f.d.renderTemplate(f.effect)
	delete(f.d.opts.data, "flag")
	return f.rendered, err
}

func (d *Driver) addCmdSpec(root *cobra.Command, arg string, cmdSpec *config.CmdSpec, run func(*cobra.Command, []string) error) {
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

			candidates := []string{}
			for _, a := range args {
				if strings.Contains(a, toComplete) {
					candidates = append(candidates, a)
				}
			}
			return candidates, cobra.ShellCompDirectiveDefault
		}
	}

	d.AddFlags(subCmd, normalizeFlags(cmdSpec.Flags))

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
