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

	eu, err = s.resolve(eu, os.Environ())
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

// resolve prepares special invokers for execution
// presently supported: gobin
func (s *Summoner) resolve(execu execUnit, environ []string) (execUnit, error) {
	if strings.HasPrefix(execu.invoker, "gobin") {
		return s.prepareGoBinExecutable(execu, environ)
	}
	return execu, nil
}

func (s *Summoner) prepareGoBinExecutable(execu execUnit, environ []string) (execUnit, error) {
	// install in OutputDir
	dest := s.opts.destination
	target := strings.Split(execu.target, "@")[0]
	targetDir := filepath.Join(dest, filepath.Dir(target))
	cmd := execCommand(execu.invoker, "-p", execu.target)
	cmd.Env = append(environ, "GOBIN="+targetDir)
	errBuf := &bytes.Buffer{}
	cmd.Stderr = errBuf
	outBuf := &bytes.Buffer{}
	cmd.Stdout = outBuf
	err := cmd.Run()

	if err != nil {
		err = errors.Wrapf(err, "executing: %s: %s", cmd.Args, errBuf)
	}
	cachePath := strings.TrimSpace(outBuf.String())

	execu.invoker = filepath.Join(targetDir, filepath.Base(cachePath))
	execu.invOpts = []string{}
	execu.target = ""

	return execu, err
}
