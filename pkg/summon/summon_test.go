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

func TestErrorOnMissingFiles(t *testing.T) {
	defer testutil.ReplaceFs()()

	box := packr.New("test box", "")
	s, _ := New(box, Filename("missing"))
	path, err := s.Summon()

	assert.NotNil(t, err)
	assert.Empty(t, path)
}

func TestMultifileInstanciation(t *testing.T) {
	defer testutil.ReplaceFs()()

	box := packr.New("test box multifile", "")

	box.AddString("text.txt", "this is a text")
	box.AddString("another.txt", "another text")

	s, _ := New(box, All(true))

	path, err := s.Summon()
	assert.Nil(t, err)
	assert.Equal(t, ".summoned", path)

	_, err = appFs.Stat(".summoned/text.txt")
	assert.Nil(t, err)

	_, err = appFs.Stat(".summoned/another.txt")
	assert.Nil(t, err)
}

func TestOneFileInstanciation(t *testing.T) {
	defer testutil.ReplaceFs()()

	a := assert.New(t)

	box := packr.New("t1", "")
	box.AddString("text.txt", "this is a text")

	// create a summoner to summon text.txt at
	s, err := New(box, Filename("text.txt"), Dest(config.OutputDir))
	a.NoError(err)

	path, err := s.Summon()
	a.NoError(err)
	a.Equal(filepath.Join(config.OutputDir, "text.txt"), path)

	bytes, err := afero.ReadFile(appFs, path)
	a.NoError(err)

	a.Equal("this is a text", string(bytes))
}
