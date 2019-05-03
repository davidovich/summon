package scaffold

import (
	"testing"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/davidovich/summon/pkg/summon"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestCreateScaffold(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		desc     string
		args     []string
		force    bool
		setup    func(afero.Fs)
		file     string
		content  string
		contains string
		wantErr  bool
	}{
		{
			desc:    "happy path",
			args:    []string{".", "example.com/my/assets", "summon"},
			file:    "go.mod",
			content: "module example.com/my/assets/summon",
		},
		{
			desc:    "output dir",
			args:    []string{"subdir", "example.com/my/assets", "summon"},
			file:    "subdir/go.mod",
			content: "module example.com/my/assets/summon",
		},
		{
			desc:    "non empty dir error",
			args:    []string{".", "example.com/my/assets", "summon"},
			wantErr: true,
			setup:   func(fs afero.Fs) { f, _ := fs.Create("sentinel"); f.Close() },
		},
		{
			desc:  "force non empty",
			args:  []string{"summon", "example.com/my/assets", "summon"},
			force: true,
			setup: func(fs afero.Fs) {
				fs.Mkdir("summon", 0644)
				f, _ := fs.Create("summon/go.mod")
				f.Write([]byte("previous content"))
				f.Close()
			},
			file:    "summon/go.mod",
			content: "module example.com/my/assets/summon",
		},
		{
			desc:    "named exe",
			args:    []string{".", "example.com/my/assets", "my-assets"},
			file:    "go.mod",
			content: "module example.com/my/assets/my-assets",
		},
		{
			desc:     "README rendered contents",
			args:     []string{".", "example.com/my/assets", "my-assets-exe"},
			file:     "README.md",
			contains: "my-assets-exe",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			defer testutil.ReplaceFs()()
			fs := summon.GetFs()

			if tC.setup != nil {
				tC.setup(fs)
			}

			err := Create(tC.args[0], tC.args[1], tC.args[2], tC.force)

			if tC.wantErr {
				assert.Error(err)
				return
			}

			bytes, err := afero.ReadFile(fs, tC.file)

			assert.NoError(err)
			if tC.content != "" {
				assert.Equal(tC.content, string(bytes))
			}
			if tC.contains != "" {
				assert.Contains(string(bytes), tC.contains)
			}
		})
	}
}
