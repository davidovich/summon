// Package summon does the heavy lifting of instanciating files or
// executing configured scripts on the user's
// machine.
//
// You can control instantiation by using Options described below.
package summon

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/afero"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/davidovich/summon/pkg/config"
)

var appFs = afero.NewOsFs()

// GetFs returns the current filesystem
func GetFs() afero.Fs {
	return appFs
}

// Summon is the main command invocation
func (d *Driver) Summon(opts ...Option) (string, error) {
	if d == nil {
		return "", fmt.Errorf("Driver cannot be nil")
	}

	err := d.Configure(opts...)
	if err != nil {
		return "", err
	}

	destination := d.opts.destination
	if destination == "-" {
		destination = ""
	}

	if d.opts.all {
		startdir := filepath.Clean(d.opts.filename)
		if d.opts.filename == "" {
			startdir = d.baseDataDir
		}

		return destination, fs.WalkDir(d.fs, startdir, makeCopyFileFun(startdir, d))
	}

	filename := filepath.Clean(d.opts.filename)
	filename = d.resolveAlias(filename)
	filename = path.Join(d.baseDataDir, filename)

	embeddedFile, err := d.fs.Open(filename)
	if err != nil {
		return "", err
	}
	stat, err := embeddedFile.Stat()
	if err != nil {
		return "", err
	}
	if stat.IsDir() {
		// User wants to extract a subdirectory
		startdir := filename
		return destination,
			fs.WalkDir(d.fs, startdir, makeCopyFileFun(startdir, d))
	}

	return d.copyOneFile(embeddedFile, filename, d.baseDataDir)
}

func makeCopyFileFun(startdir string, d *Driver) func(path string, de fs.DirEntry, _ error) error {
	return func(path string, de fs.DirEntry, _ error) error {
		if de.IsDir() {
			return nil
		}
		file, err := d.fs.Open(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(d.baseDataDir, path)
		if err != nil {
			return err
		}
		subdir, err := filepath.Rel(d.baseDataDir, startdir)
		if err != nil {
			return err
		}

		_, err = d.copyOneFile(file, rel, subdir)
		return err
	}
}

func (d *Driver) prepareTemplate() (*template.Template, error) {
	t := d.templateCtx
	var err error
	if t != nil {
		t, err = t.Clone()
		if err != nil {
			return nil, err
		}
	} else {
		t = template.New(Name)
	}

	t.Option("missingkey=zero").
		Funcs(sprig.TxtFuncMap()).
		Funcs(summonFuncMap(d))

	return t, nil
}

func executeTemplate(t *template.Template, data interface{}) (string, error) {
	buf := &bytes.Buffer{}
	err := t.Execute(buf, data)

	// The zero value for an interface is a nil interface{} which
	// has a string representation of <no value>. Strip this out.
	// https://github.com/golang/go/issues/24963
	return strings.ReplaceAll(buf.String(), "<no value>", ""), err
}

func (d *Driver) renderTemplate(tmpl string) (string, error) {
	t, err := d.prepareTemplate()
	if err != nil {
		return tmpl, err
	}

	t, err = t.Parse(tmpl)
	if err != nil {
		return tmpl, err
	}

	data := d.opts.data
	return executeTemplate(t, data)
}

func (d *Driver) resolveAlias(alias string) string {
	if resolved, ok := d.config.Aliases[alias]; ok {
		return resolved
	}

	return alias
}

func summonFuncMap(d *Driver) template.FuncMap {
	return template.FuncMap{
		"run": func(args ...string) (string, error) {
			driverCopy := Driver{
				opts:        d.opts,
				config:      d.config,
				fs:          d.fs,
				baseDataDir: d.baseDataDir,
				templateCtx: d.templateCtx,
				execCommand: d.execCommand,
				configRead:  d.configRead,
				cmdToSpec:   d.cmdToSpec,
			}
			driverCopy.opts.argsConsumed = map[int]struct{}{}
			driverCopy.opts.cobraCmd = nil
			driverCopy.opts.helpWanted.helpFlag = ""

			b := &strings.Builder{}
			err := driverCopy.Run(Ref(args[0]), Args(args[1:]...), Out(b))

			if d.opts.dryrun {
				b.WriteString("[")
				b.WriteString(args[0])
				b.WriteString(" (dry-run)]")
			}

			if d.opts.debug {
				fmt.Fprintf(os.Stderr, "Output [%s] -> `%s`...\n", args[0], b)
			}
			return strings.TrimSpace(b.String()), err
		},
		"summon": func(path string) (string, error) {
			return d.Summon(Filename(path), Dest(os.TempDir()))
		},
		"flagValue": func(flag string) (string, error) {
			for _, toRender := range d.flagsToRender {
				if toRender.name == flag {
					toRender.explicit = true
					return toRender.renderTemplate()
				}
			}
			return "", nil
		},
		"arg": func(index int, missingErrors ...string) (string, error) {
			missingError := strings.Join(missingErrors, " ")
			if d.opts.args == nil {
				return "", fmt.Errorf(missingError)
			}
			if index >= len(d.opts.args) {
				return "", fmt.Errorf("%s: index %v out of range, args: %s", missingError, index, d.opts.args)
			}

			retrieved := d.opts.args[index]
			if d.opts.argsConsumed == nil {
				d.opts.argsConsumed = map[int]struct{}{}
			}
			d.opts.argsConsumed[index] = struct{}{}
			return retrieved, nil
		},
		"args": func() []string {
			if d.opts.argsConsumed == nil {
				d.opts.argsConsumed = make(map[int]struct{}, len(d.opts.args))
			}
			for i := range d.opts.args {
				d.opts.argsConsumed[i] = struct{}{}
			}

			return d.opts.args
		},
	}
}

func (d *Driver) copyOneFile(embeddedFile fs.File, filename, root string) (string, error) {
	destination := d.opts.destination

	if !d.opts.raw {
		var err error
		filename, err = d.renderTemplate(filename)
		if err != nil {
			return "", err
		}
	}

	filename, err := filepath.Rel(root, filename)
	if err != nil {
		return "", err
	}

	var out io.Writer
	summonedFile := ""
	if destination == "-" {
		out = d.opts.out
	} else {
		summonedFile = filepath.Join(destination, filename)
		err := appFs.MkdirAll(filepath.Dir(summonedFile), os.ModePerm)
		if err != nil {
			return "", err
		}

		outf, err := appFs.Create(summonedFile)
		if err != nil {
			return "", err
		}
		defer outf.Close()
		out = outf
	}

	fileContent, err := ioutil.ReadAll(embeddedFile)
	if err != nil {
		return "", err
	}

	var rendered string
	if d.opts.raw || filename == config.ConfigFileName {
		rendered = string(fileContent)
	} else {
		rendered, err = d.renderTemplate(string(fileContent))
		if err != nil {
			return "", err
		}
	}

	_, err = io.Copy(out, bytes.NewBufferString(rendered))
	if err != nil {
		return "", err
	}

	return summonedFile, nil
}

func init() {
	testutil.SetFs = func(fs afero.Fs) { appFs = fs }
	testutil.GetFs = func() afero.Fs { return appFs }
}
