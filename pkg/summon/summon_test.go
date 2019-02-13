package summon

import (
	"path/filepath"
	"testing"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/davidovich/summon/pkg/config"
	"github.com/spf13/afero"

	"github.com/gobuffalo/packr/v2"
	"github.com/stretchr/testify/assert"
)

func TestMultifileInstanciation(t *testing.T) {
	defer testutil.ReplaceFs()()

	box := packr.New("test box", "")

	box.AddString("text.txt", "this is a text")
	box.AddString("another.txt", "another text")

	s := New(box, All(true))

	path, err := s.Summon()
	assert.Nil(t, err)
	assert.Equal(t, "", path)

	_, err = appFs.Stat("text.txt")
	assert.Nil(t, err)

	_, err = appFs.Stat("another.txt")
	assert.Nil(t, err)
}

func TestOneFileInstanciation(t *testing.T) {
	defer testutil.ReplaceFs()()

	a := assert.New(t)

	box := packr.New("t", "")
	box.AddString("text.txt", "this is a text")

	// create a summoner to summon text.txt at
	s := New(box, Filename("text.txt"), Dest(config.OutputDir))

	path, err := s.Summon()
	a.Nil(err)
	a.Equal(filepath.Join(config.OutputDir, "text.txt"), path)

	bytes, err := afero.ReadFile(appFs, path)
	a.Nil(err)

	a.Equal("this is a text", string(bytes))
}
