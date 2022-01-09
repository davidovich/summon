package summon

import "github.com/davidovich/summon/pkg/config"

// ConfigurableRunner is a runner that can be configured.
type ConfigurableRunner interface {
	Configurer
	Runner
}

// Runner allows executing configured aliases from summon.config.yaml.
type Runner interface {
	Run(opts ...Option) error
	RenderArgs(args ...interface{}) ([]string, error)
	ExecContext() (map[string]config.FlagSpec, config.Handles)
}

// Configurer allows configuring a driver from variadic options.
type Configurer interface {
	Configure(opts ...Option) error
}

// Summon is used to instantiate a real file to the filesystem.
type Summon interface {
	Summon(opts ...Option) (string, error)
}

// ConfigurableLister allows configuration and listing.
type ConfigurableLister interface {
	Configurer
	Lister
}

// Lister allows listing files in the assets.
type Lister interface {
	List(opts ...Option) ([]string, error)
}
