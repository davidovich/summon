package summon

import (
	"bytes"
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

	box := packr.New("test box", t.TempDir())
	s, _ := New(box, Filename("missing"))
	path, err := s.Summon()

	assert.NotNil(t, err)
	assert.Empty(t, path)
}

func TestMultifileInstanciation(t *testing.T) {
	defer testutil.ReplaceFs()()

	box := packr.New("test box multifile", t.TempDir())

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

	box := packr.New("t1", t.TempDir())
	box.AddString("text.txt", "this is a text")

	// create a summoner to summon text.txt at
	s, err := New(box, Filename("text.txt"), Dest(config.DefaultOutputDir))
	a.NoError(err)

	path, err := s.Summon()
	a.NoError(err)
	a.Equal(filepath.Join(config.DefaultOutputDir, "text.txt"), path)

	bytes, err := afero.ReadFile(appFs, path)
	a.NoError(err)

	a.Equal("this is a text", string(bytes))
}

func TestSubfolderHierarchy(t *testing.T) {
	defer testutil.ReplaceFs()()
	a := assert.New(t)

	box := packr.New("hierarchy", "testdata")

	// create a summoner to summon a complete hierarchy
	s, err := New(box, Filename("subdir/"), Dest("o"), JSON(`{"TemplatedName":"b", "Content":"b content"}`))
	a.NoError(err)

	path, err := s.Summon()

	a.NoError(err)
	a.Equal("o", path)

	a.True(afero.IsDir(GetFs(), "o/a"))
	a.True(afero.IsDir(GetFs(), "o/b"))
	a.True(afero.Exists(GetFs(), "o/b/b.txt"))

	bytes, err := afero.ReadFile(GetFs(), "o/b/b.txt")
	a.NoError(err)

	a.Equal("b content", string(bytes))
}

func TestSummonScenarios(t *testing.T) {
	defer testutil.ReplaceFs()()
	assert := assert.New(t)

	testCases := []struct {
		desc             string
		filename         string
		json             string
		raw              bool
		dest             string
		expectedFileName string
		expectedContent  string
		wantError        bool
	}{
		{
			desc:             "file name render",
			filename:         "renderableFileName",
			json:             `{ "FileName": "aFileName" }`,
			expectedFileName: "overridden_dir/aFileName",
			expectedContent:  "",
		},
		{
			desc:             "content render",
			filename:         "template.file",
			json:             `{ "Name": "World!" }`,
			expectedFileName: "overridden_dir/template.file",
			expectedContent:  "hello World!",
		},
		{
			desc:             "no data",
			filename:         "template.file",
			expectedFileName: "overridden_dir/template.file",
			expectedContent:  "hello ",
		},
		{
			desc:             "no data raw",
			raw:              true,
			filename:         "template.file",
			expectedFileName: "overridden_dir/template.file",
			expectedContent:  "hello {{ .Name }}",
		},
		{
			desc:             "sprig rendering",
			filename:         "sprigcontent.gotmpl",
			expectedFileName: "overridden_dir/sprigcontent.gotmpl",
			expectedContent:  "HELLO\n25",
		},
		{
			desc:             "alias",
			filename:         "a",
			expectedFileName: "overridden_dir/subdir/a/a.txt",
			expectedContent:  "this is a.txt",
		},
		{
			desc:      "error in json input",
			json:      `{ "Name": "World!"`,
			filename:  "template.file",
			wantError: true,
		},
		{
			desc:             "to stdout",
			dest:             "-",
			filename:         "a",
			expectedFileName: "",
			expectedContent:  "this is a.txt",
		},
	}

	box := packr.New("TestTemplateRendering", "testdata")

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			args := []Option{Filename(tC.filename), JSON(tC.json), Dest(tC.dest), Raw(tC.raw)}
			output := &bytes.Buffer{}
			if tC.dest == "-" {
				args = append(args, out(output))
			}
			s, err := New(box, args...)
			if tC.wantError {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			path, err := s.Summon()

			assert.Equal(tC.expectedFileName, path)
			var b []byte
			if tC.dest != "-" {
				b, _ = afero.ReadFile(appFs, path)
			} else {
				b = output.Bytes()
			}
			assert.Equal(tC.expectedContent, string(b))
		})
	}
}
