package summon

import (
	"encoding/json"
	"io"

	"github.com/davidovich/summon/internal/command"
	"github.com/davidovich/summon/pkg/config"
)

// options fir all summon commands
type options struct {
	// copy all the tree
	all bool
	// where the summoned file will land or stdout if "-"
	destination string
	// single file to instanciate
	filename string
	// show tree of files
	tree bool
	// reference to an exec config entry
	ref string
	// args to exec entry
	args []string
	// template rendering data
	data map[string]interface{}
	// out
	out io.Writer
	// raw disables template rendering
	raw bool
	// execCommand overrides the command used to run external processes
	execCommand command.ExecCommandFn
}

// Option allows specifying configuration settings
// from the user.
type Option func(*options) error

// Args captures the arguments to be passed to run.
func Args(args ...string) Option {
	return func(opts *options) error {
		opts.args = args
		return nil
	}
}

// Ref references an exec config entry.
func Ref(ref string) Option {
	return func(opts *options) error {
		opts.ref = ref
		return nil
	}
}

// All specifies to download all config files.
func All(all bool) Option {
	return func(opts *options) error {
		opts.all = all
		return nil
	}
}

// Filename sets the requested filename in the boxed data.
func Filename(filename string) Option {
	return func(opts *options) error {
		opts.filename = filename
		return nil
	}
}

// Raw disables any template rendering in assets.
func Raw(raw bool) Option {
	return func(opts *options) error {
		opts.raw = raw
		return nil
	}
}

// Dest specifies where the file(s) will be rooted.
// '-' is a special value representing stdout.
func Dest(dest string) Option {
	return func(opts *options) error {
		if dest == "" {
			return nil
		}
		opts.destination = dest
		return nil
	}
}

func out(w io.Writer) Option {
	return func(opts *options) error {
		opts.out = w
		return nil
	}
}

// ShowTree will print a pretty graph of the data tree.
func ShowTree(tree bool) Option {
	return func(opts *options) error {
		opts.tree = tree
		return nil
	}
}

// JSON configures the dictionary to use to render a templated asset.
func JSON(j string) Option {
	return func(opts *options) error {
		if j == "" {
			opts.data = map[string]interface{}{}
			return nil
		}

		var data map[string]interface{}
		err := json.Unmarshal([]byte(j), &data)

		if err != nil {
			return err
		}
		opts.data = data
		return nil
	}
}

// ExecCmd allows changing the execution function for external processes.
func ExecCmd(e command.ExecCommandFn) Option {
	return func(opts *options) error {
		opts.execCommand = e
		return nil
	}
}

// DefaultsFrom sets options from user config.
func (o *options) DefaultsFrom(conf config.Config) {
	if conf.OutputDir != "" {
		o.destination = conf.OutputDir
	}
}
