package config

const (
	// DefaultOutputDir is where the files will be instantiated
	DefaultOutputDir = ".summoned"

	// ConfigFile is the name of the summon config file
	ConfigFile = "summon.config.yaml"
)

// OutputDir is the resolved output dir
var OutputDir = DefaultOutputDir

type configFile struct {
}
