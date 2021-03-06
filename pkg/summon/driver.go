package summon

import (
	"fmt"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	"github.com/davidovich/summon/pkg/command"
	"github.com/davidovich/summon/pkg/config"
	"github.com/gobuffalo/packr/v2"
)

// Summoner is the old name for Driver, use Driver instead.
type Summoner = Driver

// Name holds the name of the driver executable. By default it is "summon"
var Name = "summon"

// Driver manages functionality of summon.
type Driver struct {
	opts        options
	Config      config.Config
	box         *packr.Box
	templateCtx *template.Template
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
			err = d.Config.Unmarshal(config)
			if err != nil {
				return err
			}
			d.opts.DefaultsFrom(d.Config)
			d.templateCtx, err = template.New(Name).
				Option("missingkey=zero").
				Funcs(sprig.TxtFuncMap()).
				Funcs(summonFuncMap(d)).
				Parse(d.Config.TemplateContext)
			if err != nil {
				return err
			}
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
