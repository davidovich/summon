package summon

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidovich/summon/pkg/command"
	"github.com/davidovich/summon/pkg/config"
	"github.com/pkg/errors"
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

	eu, err = s.resolve(eu)
	if err != nil {
		return errors.Wrapf(err, "resolving %s", eu.invoker)
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

func (s *Summoner) resolve(execu execUnit) (execUnit, error) {
	if strings.HasPrefix("gobin", execu.invoker) {
		return s.prepareGoBinExecutable(execu)
	}
	return execu, nil
}

func (s *Summoner) prepareGoBinExecutable(execu execUnit) (execUnit, error) {
	// install in OutputDir
	target := strings.Split(execu.target, "@")[0]
	targetDir := filepath.Join(s.opts.destination, filepath.Dir(target))
	cmd := execCommand(execu.invoker, execu.target)
	cmd.Env = append(os.Environ(), "GOBIN="+targetDir)
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	//fmt.Printf("executing: %s\n", cmd.Args)
	err := cmd.Run()

	if err != nil {
		err = errors.Wrapf(err, "executing: %s: %s", cmd.Args, buf)
	}

	execu.invoker = filepath.Join(targetDir, filepath.Base(target))
	execu.invOpts = []string{}
	execu.target = ""

	return execu, err
}
