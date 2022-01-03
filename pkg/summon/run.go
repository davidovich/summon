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
	invoker    string
	invOpts    string
	targets    config.ArgSliceSpec
	targetSpec config.CmdSpec
}

// Run will run executable scripts described in the summon.config.yaml file
// of the data repository module.
func (d *Driver) Run(opts ...Option) error {
	err := d.Configure(opts...)
	if err != nil {
		return err
	}

	rargs, err := d.BuildCommand()
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

		return cmd.Run()
	}

	return nil
}

func (d *Driver) BuildCommand() ([]string, error) {
	eu, err := d.findExecutor(d.opts.ref)
	if err != nil {
		return nil, err
	}

	data := d.opts.data
	// add arguments
	if data == nil {
		data = map[string]interface{}{}
	}

	data["osArgs"] = os.Args

	invOpts, err := d.renderTemplate(eu.invOpts, data)
	if err != nil {
		return nil, err
	}

	targets := make([]string, 0, len(eu.targets))
	var renderedTargets []string
	for _, t := range FlattenStrings(eu.targets) {
		rt, err := d.renderTemplate(t, data)
		if err != nil {
			return nil, err
		}

		renderedTargets = []string{rt}
		// Convert array to real array and merge
		if strings.HasPrefix(rt, "[") && strings.HasSuffix(rt, "]") {
			renderedTargets, err = shlex.Split(strings.Trim(rt, "[]"))
			if err != nil {
				return nil, err
			}
		}

		targets = append(targets, renderedTargets...)
	}

	rargs, err := shlex.Split(invOpts)
	if err != nil {
		return nil, err
	}

	rargs = append(rargs, targets...)

	unusedArgs := computeUnused(d.opts.args, d.opts.argsConsumed)
	rargs = append(rargs, unusedArgs...)

	return rargs, nil
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

// ListInvocables lists the invokers in the config file under the exec:
// key.
func (d *Driver) ListInvocables() config.Handles {
	handles := config.Handles{}

	for _, invokers := range d.Config.Exec.Invokers {
		for i, v := range invokers {
			handles[i] = v
		}
	}

	return handles
}

func (d *Driver) findExecutor(ref string) (execUnit, error) {
	eu := execUnit{}

	for ex, handles := range d.Config.Exec.Invokers {
		if c, ok := handles[ref]; ok {
			exec := strings.SplitAfterN(ex, " ", 2)
			eu.invoker = strings.TrimSpace(exec[0])
			if len(exec) == 2 {
				eu.invOpts = strings.TrimSpace(exec[1])
			}

			eu.targets, ok = c.Value.(config.ArgSliceSpec)
			if !ok {
				eu.targetSpec, ok = c.Value.(config.CmdSpec)
				if !ok {
					return execUnit{}, fmt.Errorf("config syntax error for 'exec:%s' in config %s", ref, config.ConfigFile)
				}
			}

			break
		}
	}

	if eu.invoker == "" {
		return eu, fmt.Errorf("could not find exec handle reference '%s' in config %s", d.opts.ref, config.ConfigFile)
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
