// Package config defines types and default values for summon.
package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultOutputDir is where the files will be instantiated.
	DefaultOutputDir = ".summoned"

	// ConfigFile is the name of the summon config file.
	ConfigFile = "summon.config.yaml"
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
type Handles map[string]CmdSpec

// Flags are the normalized FlagDesc
type Flags map[string]FlagSpec

// ArgSliceSpec is the basic form of args to pass to
// invoker. It can be a slice of string, or slices of strings.
type ArgSliceSpec []interface{}

// CmdSpec describes a complex command
type CmdSpec struct {
	// ExecEnvironment is the caller environment
	ExecEnvironment string
	// Cmd is the command and args that get executed in the ExecEnvironment
	Cmd ArgSliceSpec `yaml:"cmd"`
	// Args sub-arguments of current command
	Args map[string]CmdSpec `yaml:"args,omitempty"`
	// Flags of this command
	Flags map[string]FlagDesc `yaml:"flags,omitempty"`
	// Help of this command
	Help string `yaml:"help"`
	// Command to invoke to have a completion of this command
	Completion string `yaml:"completion"`
}

// FlagDesc describes a simple string flag or complex FlagSpec
type FlagDesc struct {
	Value interface{}
}

// FlagSpec is uses when you want more control on flag creation
type FlagSpec struct {
	Effect    string `yaml:"effect"`
	Shorthand string `yaml:"shorthand"`
	Help      string `yaml:"help"`
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
		value.Decode(&args)
		e.Value = args
	case yaml.MappingNode:
		cmdSpec := CmdSpec{}
		value.Decode(&cmdSpec)
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
