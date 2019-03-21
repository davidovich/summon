package summon

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/gobuffalo/packr/v2/file"
	"github.com/spf13/afero"
)

var appFs = afero.NewOsFs()

// GetFs returns the current filesystem
func GetFs() afero.Fs {
	return appFs
}

// Summon is the main comnand invocation
func (s *Summoner) Summon(opts ...Option) (string, error) {
	if s == nil {
		return "", fmt.Errorf("Sumonner cannot be nil")
	}

	err := s.Configure(opts...)
	if err != nil {
		return "", err
	}

	if s.opts.all {
		return s.opts.destination, s.box.Walk(func(path string, info file.File) error {
			_, err := s.copyOneFile(info)
			return err
		})
	}

	filename := s.resolveAlias(s.opts.filename)

	boxedFile, err := s.box.Open(filename)
	if err != nil {
		return "", err
	}
	return s.copyOneFile(boxedFile)
}

func renderTemplate(tmpl string, data *map[string]interface{}) (string, error) {
	if data == nil {
		return tmpl, nil
	}
	t, err := template.New("Summon").Parse(tmpl)
	if err != nil {
		return tmpl, err
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, data)

	return buf.String(), err
}

func (s *Summoner) resolveAlias(alias string) string {
	if resolved, ok := s.config.Aliases[alias]; ok {
		return resolved
	}

	return alias
}

func (s *Summoner) copyOneFile(boxedFile http.File) (string, error) {
	destination := s.opts.destination
	// Write the file and print it's path
	stat, _ := boxedFile.Stat()
	filename := stat.Name()

	filename, _ = renderTemplate(filename, s.opts.data)

	summonedFile := filepath.Join(destination, filename)
	err := appFs.MkdirAll(filepath.Dir(summonedFile), os.ModePerm)
	if err != nil {
		return "", err
	}

	out, err := appFs.Create(summonedFile)
	if err != nil {
		return "", err
	}
	defer out.Close()

	boxedContent, err := ioutil.ReadAll(boxedFile)

	rendered, _ := renderTemplate(string(boxedContent), s.opts.data)

	_, err = io.Copy(out, bytes.NewBufferString(rendered))
	if err != nil {
		return "", err
	}

	return summonedFile, out.Close()
}

func init() {
	testutil.SetFs = func(fs afero.Fs) { appFs = fs }
	testutil.GetFs = func() afero.Fs { return appFs }
}
