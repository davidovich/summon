package summon

import (
	"github.com/gobuffalo/packr/v2"
)

// Summoner manages functionality of summon
type Summoner struct {
	opts options
	box  *packr.Box
}

// New creates the summoner
func New(box *packr.Box, opts ...Option) *Summoner {
	s := &Summoner{
		box: box,
	}

	s.Configure(opts...)

	return s
}

// Configure is used to extract options to the object.
func (b *Summoner) Configure(opts ...Option) {
	for _, opt := range opts {
		opt(&b.opts)
	}
}
