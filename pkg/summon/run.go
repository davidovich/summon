package summon

import (
	"fmt"
	"os"
	"strings"

	"github.com/davidovich/summon/pkg/command"
	"github.com/davidovich/summon/pkg/config"
)

var execCommand = command.New

// Run will run go or executable scripts in the context of the data
func (s *Summoner) Run(opts ...Option) error {
	s.Configure(opts...)
	exec, commands, err := s.findExecutor()
	if err != nil {
		return err
	}

	finalCommand := append(commands, s.opts.args...)

	cmd := execCommand(exec, finalCommand...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (s *Summoner) findExecutor() (string, []string, error) {
	var executor string
	var commands []string

	for ex, handles := range s.config.Executables {
		if c, ok := handles[s.opts.ref]; ok {
			exec := strings.Split(ex, " ")
			executor = exec[0]
			commands = append(exec[1:], c)
			break
		}
	}

	if executor == "" {
		return "", []string{}, fmt.Errorf("could not find exec reference %s in config %s", s.opts.ref, config.ConfigFile)
	}

	return executor, commands, nil
}
