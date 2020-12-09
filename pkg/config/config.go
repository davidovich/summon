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
	Aliases         Alias                 `yaml:"aliases"`
	OutputDir       string                `yaml:"outputdir"`
	TemplateContext string                `yaml:"templates"`
	Executables     map[string]Executable `yaml:"exec"`
	Flags           map[string]string     `yaml:"flags"`
	Help            string                `yaml:"help"`
}

// ExecSpec allows unmarshaling complex subtype
type ExecSpec struct {
	Value interface{}
}

// Executable describes a handle name and invocable target.
// The ExecSpec target can be an ArgSliceSpec, or a CmdSpec
type Executable map[string]ExecSpec

// ArgSliceSpec is the basic form of args to pass to
// invoker. It is a slice of string, or slices of strings.
type ArgSliceSpec []interface{}

// CmdSpec describes a complex command
type CmdSpec struct {
	Args       []CmdSpec         `yaml:"subArgs,omitempty"`
	Flags      map[string]string `yaml:"flags,omitempty"`
	CmdArgs    ArgSliceSpec      `yaml:"cmdArgs"`
	Help       string            `yaml:"help"`
	Completion string            `yaml:"completion"`
}

// UnmarshalYAML the ExecSpec. It can be a CmdSpec or a ArgSliceSpec
func (e *ExecSpec) UnmarshalYAML(value *yaml.Node) error {
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
