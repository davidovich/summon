package summon

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/cobra"

	"github.com/davidovich/summon/pkg/command"
	"github.com/davidovich/summon/pkg/config"
)

// Summoner is the old name for Driver, use Driver instead.
type Summoner = Driver

// Name holds the name of the driver executable. By default it is "summon"
var Name = "summon"

// Driver manages functionality of summon.
type Driver struct {
	opts          options
	config        config.Config
	fs            fs.FS
	globalFlags   config.Flags
	handles       handles
	baseDataDir   string
	templateCtx   *template.Template
	execCommand   command.ExecCommandFn
	configRead    bool
	flagsToRender []*flagValue
	cmdToSpec     map[*cobra.Command]*commandSpec
}

// New creates the Driver.
func New(filesystem fs.FS, opts ...Option) (*Driver, error) {
	d := &Driver{
		fs:          filesystem,
		execCommand: command.New,
		cmdToSpec:   map[*cobra.Command]*commandSpec{},
	}

	err := fs.WalkDir(d.fs, ".", func(path string, de fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		if de.IsDir() {
			d.baseDataDir = path
			return fs.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = d.Configure(opts...)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d Driver) OutputDir() string      { return d.config.OutputDir }
func (d Driver) HideAssetsInHelp() bool { return d.config.HideAssetsInHelp }

// Configure is used to extract options and customize the summon.Driver.
func (d *Driver) Configure(opts ...Option) error {
	if d == nil {
		return fmt.Errorf("driver cannot be nil")
	}
	if !d.configRead {
		// try to find a config file in the embedded assets filesystem
		configFile, err := d.fs.Open(path.Join(d.baseDataDir, config.ConfigFileName))
		if err == nil {
			defer configFile.Close()
			config, err := io.ReadAll(configFile)
			if err != nil {
				return err
			}
			err = d.config.Unmarshal(config)
			if err != nil {
				return err
			}
			d.opts.DefaultsFrom(d.config)
			d.templateCtx, err = template.New(Name).
				Option("missingkey=zero").
				Funcs(sprig.TxtFuncMap()).
				Funcs(summonFuncMap(d)).
				Parse(d.config.TemplateContext)
			if err != nil {
				return err
			}
			// prime execContext cache
			_, _, err = d.execContext()
			if err != nil {
				return err
			}

			d.configRead = true
		}
	}

	for _, opt := range opts {
		err := opt(&d.opts)
		if err != nil {
			return err
		}
	}

	if d.opts.destination == "" {
		d.opts.destination = config.DefaultOutputDir
	}

	if d.opts.out == nil {
		d.opts.out = os.Stdout
	}

	if d.opts.execCommand != nil {
		d.execCommand = d.opts.execCommand
	}

	// add arguments
	if d.opts.data == nil {
		d.opts.data = map[string]interface{}{}
	}

	d.opts.data["osArgs"] = os.Args

	return nil
}

// manage json manually
type jsonValue struct {
	d             Configurer
	cmd           *cobra.Command
	isFile        bool
	otherValueSet *bool
	valueSet      bool
	json          string
}

func (jv *jsonValue) Set(s string) error {
	jv.json = s
	if *jv.otherValueSet {
		return fmt.Errorf("--json and --json-file are mutually exclusive")
	}
	jv.valueSet = true
	if jv.isFile {
		var j []byte
		var err error
		if s == "-" {
			j, err = ioutil.ReadAll(jv.cmd.InOrStdin())
		} else {
			j, err = ioutil.ReadFile(s)
		}
		if err != nil {
			return err
		}

		jv.json = string(j)
	}
	return jv.d.Configure(JSON(&jv.json))
}

func (jv *jsonValue) String() string {
	return jv.json
}

func (jv *jsonValue) Type() string { return "string" }

func (d *Driver) RegisterFlags(runRoot *cobra.Command) {
	json := &jsonValue{d: d, cmd: runRoot.Root()}
	jsonFile := &jsonValue{d: d, cmd: runRoot.Root(), isFile: true, otherValueSet: &json.valueSet}
	json.otherValueSet = &jsonFile.valueSet

	runRoot.Root().PersistentFlags().Var(json, "json", "json to use to render template")
	runRoot.Root().PersistentFlags().Var(jsonFile, "json-file", "json file to use to render template, with '-' for stdin")

	runRoot.Root().Flags().BoolVarP(&d.opts.debug, "debug", "d", false, "print debug info on stderr")
	runRoot.Flags().BoolVarP(&d.opts.dryrun, "dry-run", "n", false, "only show what would be executed")
}
