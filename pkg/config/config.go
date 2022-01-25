// Package config defines types and default values for summon.
package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultOutputDir is where the files will be instantiated.
	DefaultOutputDir = ".summoned"

	// ConfigFileName is the name of the summon config file.
	ConfigFileName = "summon.config.yaml"
)

// Alias gives a shortcut to a name in data.
type Alias map[string]string

// Config is the summon config
type Config struct {
	Version         int
	Aliases         Alias       `yaml:"aliases"`
	OutputDir       string      `yaml:"outputdir"`
	TemplateContext string      `yaml:"templates"`
	Exec            ExecContext `yaml:"exec"`
	Help            string      `yaml:"help"`
}

// ExecContext houses execution environments and global flags
type ExecContext struct {
	ExecEnv     map[string]HandlesDesc `yaml:"environments"`
	GlobalFlags map[string]FlagDesc    `yaml:"flags"`
}

// ExecDesc allows unmarshaling complex subtype
type ExecDesc struct {
	Value interface{}
}

// HandlesDesc describes a handle name and invocable target.
// The ExecDesc target can be an ArgSliceSpec, or a CmdDesc
type HandlesDesc map[string]ExecDesc

// Flags are the normalized FlagDesc
type Flags map[string]*FlagSpec

// ArgSliceSpec is the basic form of args to pass to
// invoker. It can be a slice of string, or slices of strings.
type ArgSliceSpec []interface{}

// CmdDesc describes a polymorphic Cmd in the config file.
// Its SubCmd is an ExecDesc so it can be an ArgsSliceSpec or a CmdSpec
// Its Flags can be a one line string flag or a FlagSpec
type CmdDesc struct {
	Args       ArgSliceSpec        `yaml:"args"`
	SubCmd     map[string]ExecDesc `yaml:"subCmd,omitempty"`
	Flags      map[string]FlagDesc `yaml:"flags,omitempty"`
	Help       string              `yaml:"help,omitempty"`
	Completion string              `yaml:"completion,omitempty"`
	Hidden     bool                `yaml:"hidden,omitempty"`
}

// FlagDesc describes a simple string flag or complex FlagSpec
type FlagDesc struct {
	Value interface{}
}

// FlagSpec is used when you want more control on flag creation
type FlagSpec struct {
	// Effect contains the value that will be assigned to the flag, after
	// template rendering, if needed
	Effect string `yaml:"effect"`
	// Shorthand (one letter) for the flag
	Shorthand string `yaml:"shorthand"`
	// Default value if the value is not provided by the user
	Default string `yaml:"default"`
	// Help for the flag
	Help string `yaml:"help"`
	// Explicit is used to control if the flag is added automatically or not to
	// the command-line. If Explicit is true, the flag will not be automatically
	// added (it can be positionned in the command-line with the flagValue template
	// function). The default is to add the rendered flag on the command line (implicit).
	// Note that using the {{ flagValue "my-flag" }} in a template makes the Flag Explicit.
	Explicit bool `yaml:"explicit"`
}

// UnmarshalYAML the FlagSpec. It can be a String or a Flag
func (e *FlagDesc) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var args string
		value.Decode(&args)
		e.Value = args
	case yaml.MappingNode:
		flag := FlagSpec{}
		value.Decode(&flag)
		e.Value = flag
	default:
		return &yaml.TypeError{
			Errors: []string{fmt.Sprintf("cannot unmarshal %v, content: %v", value.Tag, value.Content)},
		}
	}
	return nil
}

// UnmarshalYAML the ExecDesc. It can be a CmdSpec or a ArgSliceSpec
func (e *ExecDesc) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.SequenceNode:
		args := ArgSliceSpec{}
		err := value.Decode(&args)
		if err != nil {
			return &yaml.TypeError{
				Errors: []string{fmt.Sprintf("cannot unmarshal %v, content: %+v", value.Tag, value.Content)},
			}
		}
		e.Value = args
	case yaml.MappingNode:
		cmdDesc := CmdDesc{}
		err := value.Decode(&cmdDesc)
		if err != nil {
			return &yaml.TypeError{
				Errors: []string{fmt.Sprintf("cannot unmarshal %v on line: %d, colunm: %d, content: %+v", value.Tag, value.Line, value.Column, cmdDesc)},
			}
		}
		e.Value = cmdDesc
	default:
		return &yaml.TypeError{
			Errors: []string{fmt.Sprintf("cannot unmarshal %v, content: %v", value.Tag, value.Content)},
		}
	}
	return nil
}

// Unmarshal hidrates the config from config bytes.
func (c *Config) Unmarshal(config []byte) error {
	return yaml.Unmarshal(config, c)
}
