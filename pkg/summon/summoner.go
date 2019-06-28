package summon

import (
	"os"

	"github.com/davidovich/summon/pkg/config"
	"github.com/gobuffalo/packr/v2"
)

// Driver manages functionality of summon
type Driver struct {
	opts       options
	config     config.Config
	box        *packr.Box
	configRead bool
}

// New creates the summoner
func New(box *packr.Box, opts ...Option) (*Driver, error) {
	s := &Driver{
		box: box,
	}

	err := s.Configure(opts...)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Configure is used to extract options to the object.
func (b *Driver) Configure(opts ...Option) error {
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

	if b.opts.destination == "" {
		b.opts.destination = config.DefaultOutputDir
	}

	if b.opts.out == nil {
		b.opts.out = os.Stdout
	}

	return nil
}
