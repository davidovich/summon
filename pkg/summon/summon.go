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
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/gobuffalo/packr/v2/file"
	"github.com/spf13/afero"

	"github.com/davidovich/summon/internal/testutil"
)

var appFs = afero.NewOsFs()

// GetFs returns the current filesystem
func GetFs() afero.Fs {
	return appFs
}

// Summon is the main comnand invocation
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
		return destination, d.box.Walk(func(path string, info file.File) error {
			_, err := d.copyOneFile(info, "")
			return err
		})
	}

	filename := filepath.Clean(d.opts.filename)
	filename = d.resolveAlias(filename)

	// User wants to extract a subdirectory
	if d.box.HasDir(filename) {
		return destination,
			d.box.WalkPrefix(filename, func(path string, info file.File) error {
				_, err := d.copyOneFile(info, filename)
				return err
			})
	}

	boxedFile, err := d.box.Open(filename)
	if err != nil {
		return "", err
	}
	return d.copyOneFile(boxedFile, "")
}

func (d *Driver) renderTemplate(tmpl string, data map[string]interface{}) (string, error) {
	t := d.templateCtx
	var err error
	if t != nil {
		t, err = t.Clone()
		if err != nil {
			return tmpl, err
		}
	} else {
		t = template.New(Name).
			Option("missingkey=zero").
			Funcs(sprig.TxtFuncMap()).
			Funcs(summonFuncMap(d))
	}

	t, err = t.Parse(tmpl)
	if err != nil {
		return tmpl, err
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, data)

	// The zero value for an interface is a nil interface{} which
	// has a string representation of <no value>. Strip this out.
	// https://github.com/golang/go/issues/24963
	return strings.ReplaceAll(buf.String(), "<no value>", ""), err
}

func (d *Driver) resolveAlias(alias string) string {
	if resolved, ok := d.config.Aliases[alias]; ok {
		return resolved
	}

	return alias
}

func summonFuncMap(d *Driver) template.FuncMap {
	return template.FuncMap{
		"summon": func(path string) (string, error) {
			return d.Summon(Filename(path), Dest(os.TempDir()))
		},
	}
}

func (d *Driver) copyOneFile(boxedFile http.File, rootDir string) (string, error) {
	destination := d.opts.destination
	// Write the file and print it'd path
	stat, err := boxedFile.Stat()
	if err != nil {
		return "", err
	}
	filename := stat.Name()

	if !d.opts.raw {
		filename, err = d.renderTemplate(filename, d.opts.data)
		if err != nil {
			return "", err
		}
	}

	filename, err = filepath.Rel(rootDir, filename)
	if err != nil {
		return "", err
	}

	var out io.Writer
	summonedFile := ""
	if destination == "-" {
		out = d.opts.out
	} else {
		summonedFile = filepath.Join(destination, filename)
		err = appFs.MkdirAll(filepath.Dir(summonedFile), os.ModePerm)
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

	boxedContent, err := ioutil.ReadAll(boxedFile)

	var rendered string
	if d.opts.raw {
		rendered = string(boxedContent)
	} else {
		rendered, err = d.renderTemplate(string(boxedContent), d.opts.data)
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
