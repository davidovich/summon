package summon

import (
	"github.com/spf13/cobra"
)

// ConfigurableRunner is a runner that can be configured.
type ConfigurableRunner interface {
	Configurer
	Runner
}

// Runner allows executing configured aliases from summon.config.yaml.
type Runner interface {
	Run(opts ...Option) error
	ConstructCommandTree(*cobra.Command, bool) (*cobra.Command, error)
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

// Prompter allows prompting the user.
type Prompter interface {
	NewPrompt(userPrompt string)
	Choose(choices []string) (string, error)
	Input(defaultVal string) (string, error)
}
