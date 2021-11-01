package cmd

import (
	"bytes"
	"testing"

	"github.com/davidovich/summon/pkg/summon"
	"github.com/stretchr/testify/assert"
)

func TestCompletionCommand(t *testing.T) {
	s, _ := summon.New(runCmdTestFS)

	cmd := newCompletionCmd(s)

	b := &bytes.Buffer{}

	cmd.SetOutput(b)
	cmd.Execute()

	assert.Contains(t, b.String(), "completion_summon.config.yaml")
}
