package summon

import (
	"github.com/gobuffalo/packr/v2"
)

type options struct {
	// copy all the tree
	all bool
	// where the summoned file will land
	destination string
	// Box virtual filesystem
	box *packr.Box
}

// Option allows specifying configuration settings
// from the user
type Option func(*options)

// All specifies to download all config files
func All(all bool) Option {
	return func(opts *options) {
		opts.all = all
	}
}

// Dest specifies where the file(s) will be rooted
func Dest(dest string) Option {
	return func(opts *options) {
		opts.destination = dest
	}
}

// Box sets the box with the virtual file system
func Box(box *packr.Box) Option {
	return func(opts *options) {
		opts.box = box
	}
}
