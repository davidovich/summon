package summon

import (
	"github.com/davidovich/summon/pkg/config"
	"github.com/gobuffalo/packr/v2"
)

// Summoner manages functionality of summon
type Summoner struct {
	opts   options
	config config.Config
	box    *packr.Box
}

// New creates the summoner
func New(box *packr.Box, opts ...Option) *Summoner {
	s := &Summoner{
		box: box,
	}

	s.Configure(opts...)

	return s
}

// Configure is used to extract options to the object.
func (b *Summoner) Configure(opts ...Option) {
	if b.opts.destination == "" {
		b.opts.destination = config.DefaultOutputDir
	}

	// try to find a config file in the box
	config, err := b.box.Find(config.ConfigFile)
	if err == nil {
		b.config.Unmarshal(config)
		b.opts.DefaultsFrom(b.config)
	}

	for _, opt := range opts {
		opt(&b.opts)
	}
}
