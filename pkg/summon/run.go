package summon

import (
	"fmt"
	"os"
	"strings"

	"github.com/davidovich/summon/pkg/config"
)

type execUnit struct {
	invoker string
	invOpts []string
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

	args := eu.invOpts
	if eu.target != "" {
		args = append(args, eu.target)
	}
	args = append(args, d.opts.args...)

	cmd := d.execCommand(eu.invoker, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
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
			exec := strings.Split(ex, " ")
			eu.invoker = exec[0]
			eu.invOpts = exec[1:]
			eu.target = c
			break
		}
	}

	if eu.invoker == "" {
		return eu, fmt.Errorf("could not find exec handle reference %s in config %s", d.opts.ref, config.ConfigFile)
	}

	return eu, nil
}
