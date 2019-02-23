package config

import "gopkg.in/yaml.v2"

const (
	// DefaultOutputDir is where the files will be instantiated
	DefaultOutputDir = ".summoned"

	// ConfigFile is the name of the summon config file
	ConfigFile = "summon.config.yaml"
)

// OutputDir is the resolved output dir
var OutputDir = DefaultOutputDir

// Alias gives a shortcut to a name in data
type Alias map[string]string

// Executable describes a Name and an invocable target
type Executable map[string]string

// Config is the summon config
type Config struct {
	Version     int
	Aliases     Alias                 `yaml:"aliases"`
	OutputDir   string                `yaml:"outputdir"`
	Executables map[string]Executable `yaml:"exec"`
}

// Unmarshal hidrates the config from config bytes
func (c *Config) Unmarshal(config []byte) error {
	return yaml.Unmarshal(config, c)
}
