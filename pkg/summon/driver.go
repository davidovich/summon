package summon

import (
	"fmt"
	"os"

	"github.com/davidovich/summon/pkg/command"
	"github.com/davidovich/summon/pkg/config"
	"github.com/gobuffalo/packr/v2"
)

// Summoner is the old name for Driver, use Driver instead.
type Summoner = Driver

// Driver manages functionality of summon.
type Driver struct {
	opts        options
	config      config.Config
	box         *packr.Box
	execCommand command.ExecCommandFn
	configRead  bool
}

// New creates the Driver.
func New(box *packr.Box, opts ...Option) (*Driver, error) {
	d := &Driver{
		box:         box,
		execCommand: command.New,
	}

	err := d.Configure(opts...)
	if err != nil {
		return nil, err
	}

	return d, nil
}

// Configure is used to extract options and customize the summon.Driver.
func (d *Driver) Configure(opts ...Option) error {
	if d == nil {
		return fmt.Errorf("driver cannot be nil")
	}
	if !d.configRead {
		// try to find a config file in the box
		config, err := d.box.Find(config.ConfigFile)
		if err == nil {
			err = d.config.Unmarshal(config)
			if err != nil {
				return err
			}
			d.opts.DefaultsFrom(d.config)
			d.configRead = true
		}
	}

	for _, opt := range opts {
		err := opt(&d.opts)
		if err != nil {
			return err
		}
	}

	if d.opts.destination == "" {
		d.opts.destination = config.DefaultOutputDir
	}

	if d.opts.out == nil {
		d.opts.out = os.Stdout
	}

	if d.opts.execCommand != nil {
		d.execCommand = d.opts.execCommand
	}

	return nil
}
