package summon

// Interface for summon
type ConfigurableRunner interface {
	Configurer
	Runner
}

type Runner interface {
	Run(opts ...Option) error
	ListInvocables() []string
}

type Configurer interface {
	Configure(opts ...Option) error
}

type Summoner interface {
	Summon(opts ...Option) (string, error)
}

type ConfigurableLister interface {
	Configurer
	Lister
}

type Lister interface {
	List(opts ...Option) ([]string, error)
}

type ListerRunnerSummoner interface {
	Configurer
	Lister
	Runner
	Summoner
}
