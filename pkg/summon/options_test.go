package summon

import (
	"testing"

	"github.com/gobuffalo/packr/v2"
	"github.com/stretchr/testify/assert"
)

func TestBoxedConfig(t *testing.T) {
	box := packr.New("testBoxedConfig", "testdata")

	s := New(box)

	assert.Equal(t, "overriden_dir", s.opts.destination)
}
