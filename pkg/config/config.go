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

// ExecContext houses invokers and global flags
type ExecContext struct {
	Invokers    map[string]HandlesDesc `yaml:"invokers"`
	GlobalFlags map[string]FlagDesc    `yaml:"flags"`
}

// ExecDesc allows unmarshaling complex subtype
type ExecDesc struct {
	Value interface{}
}

// HandlesDesc describes a handle name and invocable target.
// The ExecDesc target can be an ArgSliceSpec, or a CmdSpec
type HandlesDesc map[string]ExecDesc

// Handles are the normalized version of the configs HandleDesc
type Handles map[string]*CmdSpec

// Flags are the normalized FlagDesc
type Flags map[string]*FlagSpec

// ArgSliceSpec is the basic form of args to pass to
// invoker. It can be a slice of string, or slices of strings.
type ArgSliceSpec []interface{}

// CmdSpec describes a complex command
type CmdSpec struct {
	// ExecEnvironment is the caller environment (docker, bash, python)
	ExecEnvironment string
	// Cmd is the command and args that get executed in the ExecEnvironment
	Cmd ArgSliceSpec `yaml:"cmd"`
	// Args sub-arguments of current command
	Args map[string]*CmdSpec `yaml:"args,omitempty"`
	// Flags of this command
	Flags map[string]FlagDesc `yaml:"flags,omitempty"`
	// Help of this command
	Help string `yaml:"help"`
	// Command to invoke to have a completion of this command
	Completion string `yaml:"completion"`
	// Hidden hides the command from help
	Hidden bool `yaml:"hidden"`
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
	// function). The default is to add the rendered flag on the command line (implicit).  Note that using the {{ flagValue "my-flag" }} in a template makes the Flag Explicit.
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
		cmdSpec := CmdSpec{}
		err := value.Decode(&cmdSpec)
		if err != nil {
			return &yaml.TypeError{
				Errors: []string{fmt.Sprintf("cannot unmarshal %v on line: %d, colunm: %d, content: %+v", value.Tag, value.Line, value.Column, cmdSpec)},
			}
		}
		e.Value = cmdSpec
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
