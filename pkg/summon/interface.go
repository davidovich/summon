package summon

// Interface for summon
type Interface interface {
	Configure(opts ...Option) error
	Summon(opts ...Option) (string, error)
	List(opts ...Option) ([]string, error)
	Run(opts ...Option) error
}
