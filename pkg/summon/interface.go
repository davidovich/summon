package summon

// Interface for summon
type Interface interface {
	Summon(opts ...Option) error
	List(opts ...Option) error
	Run(opts ...Option) error
}
