package summon

import "github.com/davidovich/summon/pkg/config"

// options fir all summon commands
type options struct {
	// copy all the tree
	all bool
	// where the summoned file will land
	destination string
	// single file to instanciate
	filename string
	// show tree of files
	tree bool
	// reference to an exec config entry
	ref string

	args []string
}

// Option allows specifying configuration settings
// from the user
type Option func(*options) error

// Args captures the arguments to be passed to run
func Args(args ...string) Option {
	return func(opts *options) error {
		opts.args = args
		return nil
	}
}

// Ref references an exec config entry
func Ref(ref string) Option {
	return func(opts *options) error {
		opts.ref = ref
		return nil
	}
}

// All specifies to download all config files
func All(all bool) Option {
	return func(opts *options) error {
		opts.all = all
		return nil
	}
}

// Filename sets the reuqested filename in the boxed data
func Filename(filename string) Option {
	return func(opts *options) error {
		opts.filename = filename
		return nil
	}
}

// Dest specifies where the file(s) will be rooted
func Dest(dest string) Option {
	return func(opts *options) error {
		opts.destination = dest
		return nil
	}
}

// ShowTree will print a pretty graph of the data tree
func ShowTree(tree bool) Option {
	return func(opts *options) error {
		opts.tree = tree
		return nil
	}
}

// DefaultsFrom sets options from user config
func (o *options) DefaultsFrom(conf config.Config) {
	if conf.OutputDir != "" {
		o.destination = conf.OutputDir
	}
}
