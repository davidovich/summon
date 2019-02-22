package summon

import "os/exec"

var execCommand = exec.Command

// Run will run go or executable scripts in the context of the data
func (s *Summoner) Run(opts ...Option) error {
	var executor string
	var command string

	for k, v := range s.config.Executables {
		if c, ok := v[s.opts.ref]; ok {
			executor = k
			command = c
		}
	}
	_ = executor
	_ = command
	//cmd := execCommand(executor, append([]string{command}, s.opts.args))
	return nil // cmd.Run()
}
