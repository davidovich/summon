package summon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoxedConfig(t *testing.T) {
	s, _ := New(summonTestFS)

	assert.Equal(t, "overridden_dir", s.opts.destination)
}

func TestOptions(t *testing.T) {
	configure := func(o *options, opts ...Option) {
		for _, opt := range opts {
			err := opt(o)
			assert.NoError(t, err)
		}
	}

	t.Run("dry-run-debug", func(t *testing.T) {
		config := options{}
		configure(&config, DryRun(true), Debug(true))

		assert.Equal(t, true, config.dryrun)
		assert.Equal(t, true, config.debug)
	})

	t.Run("dest", func(t *testing.T) {
		config := options{}
		configure(&config, Dest("a"))
		assert.Equal(t, "a", config.destination)
	})

	t.Run("empty-dest", func(t *testing.T) {
		config := options{destination: "already-set"}
		configure(&config, Dest(""))
		assert.Equal(t, "already-set", config.destination)
	})

	t.Run("json", func(t *testing.T) {
		o := &options{}
		err := JSON(nil)(o)
		assert.Error(t, err)

		str := ""
		err = JSON(&str)(o)
		assert.NoError(t, err)
		assert.NotNil(t, o.data)

		// invalid json
		str = `f{k`
		err = JSON(&str)(o)
		assert.Error(t, err)
	})
}
