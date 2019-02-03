package summon

import (
	"github.com/gobuffalo/packr/v2"
)

// BaseSummoner implements options and options setup
type BaseSummoner struct {
	opts options
	box  *packr.Box
}

// Summoner manages functionality of summon
type Summoner struct {
	BaseSummoner
}

// New creates the summoner
func New(box *packr.Box, opts ...Option) *Summoner {
	s := &Summoner{
		BaseSummoner{
			box: box,
		},
	}

	s.Configure(opts...)

	return s
}

// Configure is used to extract options to the object.
func (b *BaseSummoner) Configure(opts ...Option) {
	for _, opt := range opts {
		opt(&b.opts)
	}
}
