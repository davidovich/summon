package summon

import (
	"fmt"
	"os"
	"strings"

	"github.com/davidovich/summon/pkg/command"
	"github.com/davidovich/summon/pkg/config"
)

var execCommand = command.New

type execUnit struct {
	invoker string
	invOpts []string
	target  string
}

// Run will run go or executable scripts in the context of the data
func (s *Summoner) Run(opts ...Option) error {
	s.Configure(opts...)

	eu, err := s.findExecutor()
	if err != nil {
		return err
	}

	args := eu.invOpts
	if eu.target != "" {
		args = append(args, eu.target)
	}
	args = append(args, s.opts.args...)

	cmd := execCommand(eu.invoker, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (s *Summoner) findExecutor() (execUnit, error) {
	eu := execUnit{}

	for ex, handles := range s.config.Executables {
		if c, ok := handles[s.opts.ref]; ok {
			exec := strings.Split(ex, " ")
			eu.invoker = exec[0]
			eu.invOpts = exec[1:]
			eu.target = c
			break
		}
	}

	if eu.invoker == "" {
		return eu, fmt.Errorf("could not find exec handle reference %s in config %s", s.opts.ref, config.ConfigFile)
	}

	return eu, nil
}
