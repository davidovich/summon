package summon

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/shlex"
	"gopkg.in/alessio/shellescape.v1"

	"github.com/davidovich/summon/pkg/config"
)

type execUnit struct {
	invoker string
	invOpts string
	target  string
}

// Run will run executable scripts described in the summon.config.yaml file
// of the data repository module.
func (d *Driver) Run(opts ...Option) error {
	d.Configure(opts...)

	eu, err := d.findExecutor()
	if err != nil {
		return err
	}

	args := []string{eu.invOpts}
	if eu.target != "" {
		args = append(args, eu.target)
	}
	args = append(args, d.opts.args...)

	// render arg templates
	rargs := make([]string, 0, len(args))
	for _, a := range args {
		if a == "" {
			continue
		}
		rarg, err := d.renderTemplate(a, d.opts.data)
		if err != nil {
			return err
		}

		allrargs, err := shlex.Split(rarg)
		if err != nil {
			return err
		}
		rargs = append(rargs, allrargs...)
	}

	cmd := d.execCommand(eu.invoker, rargs...)

	if d.opts.debug || d.opts.dryrun {
		msg := "Executing"
		if d.opts.dryrun {
			msg = "Would execute"
		}
		fmt.Fprintf(os.Stderr, "%s `%s`...\n", msg, cmd.Args)
	}

	if !d.opts.dryrun {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	}
	return nil
}

// ListInvocables lists the invocables in the config file under the exec:
// key.
func (d *Driver) ListInvocables() []string {
	invocables := []string{}

	for _, handles := range d.config.Executables {
		for i := range handles {
			invocables = append(invocables, i)
		}
	}

	return invocables
}

func (d *Driver) findExecutor() (execUnit, error) {
	eu := execUnit{}

	for ex, handles := range d.config.Executables {
		if c, ok := handles[d.opts.ref]; ok {
			exec := strings.SplitAfterN(ex, " ", 2)
			eu.invoker = strings.TrimSpace(exec[0])
			if len(exec) == 2 {
				eu.invOpts = strings.TrimSpace(exec[1])
			}
			if c != "" {
				eu.target = shellescape.Quote(c)
			}

			break
		}
	}

	if eu.invoker == "" {
		return eu, fmt.Errorf("could not find exec handle reference %s in config %s", d.opts.ref, config.ConfigFile)
	}

	return eu, nil
}
