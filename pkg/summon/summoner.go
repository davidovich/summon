package summon

import (
	"github.com/davidovich/summon/pkg/config"
	"github.com/gobuffalo/packr/v2"
)

// Summoner manages functionality of summon
type Summoner struct {
	opts       options
	config     config.Config
	box        *packr.Box
	configRead bool
}

// New creates the summoner
func New(box *packr.Box, opts ...Option) (*Summoner, error) {
	s := &Summoner{
		box: box,
	}

	err := s.Configure(opts...)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Configure is used to extract options to the object.
func (b *Summoner) Configure(opts ...Option) error {
	if b.opts.destination == "" {
		b.opts.destination = config.DefaultOutputDir
	}

	if !b.configRead {
		// try to find a config file in the box
		config, err := b.box.Find(config.ConfigFile)
		if err == nil {
			err = b.config.Unmarshal(config)
			if err != nil {
				return err
			}
			b.opts.DefaultsFrom(b.config)
			b.configRead = true
		}
	}

	for _, opt := range opts {
		err := opt(&b.opts)
		if err != nil {
			return err
		}
	}

	return nil
}
