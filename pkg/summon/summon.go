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
	"os"
	"path"
	"path/filepath"

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

func (d *Driver) resolveAlias(alias string) string {
	if resolved, ok := d.config.Aliases[alias]; ok {
		return resolved
	}

	return alias
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

	fileContent, err := io.ReadAll(embeddedFile)
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
