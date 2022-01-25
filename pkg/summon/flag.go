package summon

import (
	"github.com/davidovich/summon/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type flagValue struct {
	d            *Driver
	name         string
	effect       string
	userValue    string
	rendered     string
	explicit     bool
	initializing bool
}

func (f *flagValue) Set(s string) error {
	if f.d.flagsToRender == nil {
		f.d.flagsToRender = []*flagValue{}
	}
	var found bool
	for _, fv := range f.d.flagsToRender {
		if fv.name == f.name {
			found = true
			break
		}
	}
	if !found {
		f.d.flagsToRender = append(f.d.flagsToRender, f)
	}
	f.userValue = s
	return nil
}

// String returns the current value
func (f *flagValue) String() string {
	if f.initializing && f.name == "help" { // fool cobra as it is very insistant on having a help
		return "false"
	}

	return f.userValue
}

func (f *flagValue) Type() string {
	if f.initializing && f.name == "help" { // fool cobra as it is very insistant on having a help
		return "bool"
	}
	return "string"
}

func (f *flagValue) renderTemplate() (string, error) {
	if f.rendered != "" {
		return f.rendered, nil
	}
	var err error
	f.d.opts.data["flag"] = f.userValue
	f.rendered, err = f.d.renderTemplate(f.effect)
	delete(f.d.opts.data, "flag")
	return f.rendered, err
}

func (d *Driver) AddFlags(cmd *cobra.Command, flags config.Flags, global bool) {
	for f, flagSpec := range flags {
		d.AddFlag(cmd, f, flagSpec, global)
	}
}

func (d *Driver) AddFlag(cmd *cobra.Command, name string, flagSpec *config.FlagSpec, global bool) *flagValue {
	v := &flagValue{
		name:     name,
		d:        d,
		effect:   flagSpec.Effect,
		explicit: flagSpec.Explicit,
	}
	var flag *pflag.Flag
	if global {
		flag = cmd.PersistentFlags().VarPF(v, name, flagSpec.Shorthand, flagSpec.Help)
	} else {
		flag = cmd.Flags().VarPF(v, name, flagSpec.Shorthand, flagSpec.Help)
	}
	flag.NoOptDefVal = flagSpec.Default
	return v
}
