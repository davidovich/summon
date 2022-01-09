package summon

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/google/shlex"

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
