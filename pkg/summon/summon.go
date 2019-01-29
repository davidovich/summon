package summon

// Summoner manages functionality off summon
type Summoner struct {
	opts options
}

// New creates the summoner
func New(opts ...Option) *Summoner {
	var s Summoner

	return s.configure(opts...)
}

func (s *Summoner) configure(opts ...Option) *Summoner {
	for _, opt := range opts {
		opt(&s.opts)
	}
}
