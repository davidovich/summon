package summon

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gobuffalo/packr/v2/file"
	"github.com/spf13/afero"
)

// AppFs abstracts away the filesystem
var AppFs = afero.NewOsFs()

// Summon is the main comnand invocation
func (s *Summoner) Summon(opts ...Option) (string, error) {
	s.Configure(opts...)

	if s.opts.all {
		return s.opts.destination, s.box.Walk(func(path string, info file.File) error {
			_, err := s.copyOneFile(info, s.opts.destination)
			return err
		})
	}

	boxedFile, err := s.box.Open(s.opts.filename)
	if err != nil {
		return "", err
	}
	return s.copyOneFile(boxedFile, s.opts.destination)
}

func (s *Summoner) copyOneFile(boxedFile http.File, destination string) (string, error) {
	// Write the file and print it's path
	stat, _ := boxedFile.Stat()
	filename := stat.Name()
	summonedFile := filepath.Join(destination, filename)
	err := AppFs.MkdirAll(filepath.Dir(summonedFile), os.ModePerm)
	if err != nil {
		return "", err
	}

	out, err := AppFs.Create(summonedFile)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, boxedFile)
	if err != nil {
		return "", err
	}

	return summonedFile, out.Close()
}
