package summon

// Interface for summon
type Interface interface {
	Summon(opts ...Option) (string, error)
	List(opts ...Option) ([]string, error)
	Run(opts ...Option) error
}
