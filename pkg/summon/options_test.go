package summon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoxedConfig(t *testing.T) {
	s, _ := New(summonTestFS)

	assert.Equal(t, "overridden_dir", s.opts.destination)
}
