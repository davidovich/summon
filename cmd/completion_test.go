package cmd

import (
	"bytes"
	"testing"

	"github.com/davidovich/summon/pkg/summon"
	"github.com/gobuffalo/packr/v2"
	"github.com/stretchr/testify/assert"
)

func TestCompletionCommand(t *testing.T) {
	box := packr.New("testCompletion", "testdata/plain")

	s, _ := summon.New(box)

	cmd := newCompletionCmd(s)

	b := &bytes.Buffer{}

	cmd.SetOutput(b)
	cmd.Execute()

	assert.Contains(t, b.String(), "completion_summon.config.yaml")
}
