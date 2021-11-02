package main

import (
	"testing"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/davidovich/summon/pkg/summon"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestScaffolder(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		desc    string
		args    []string
		setup   func(afero.Fs)
		file    string
		content string
		wantErr bool
	}{
		{
			desc:    "happy path",
			args:    []string{"init", "example.com/my/assets"},
			file:    "go.mod",
			content: "module example.com/my/assets\n",
		},
		{
			desc:    "non empty dir error",
			args:    []string{".", "example.com/my/assets", "summon"},
			wantErr: true,
			setup:   func(fs afero.Fs) { f, _ := fs.Create("sentinel"); f.Close() },
		},
		{
			desc:    "missing param",
			args:    []string{"init"},
			wantErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			defer testutil.ReplaceFs()()
			fs := summon.GetFs()

			if tC.setup != nil {
				tC.setup(fs)
			}

			cmd := newMainCmd()
			cmd.SetArgs(tC.args)
			errCode := execute(cmd)

			if tC.wantErr {
				assert.Equal(1, errCode)
				return
			}
			assert.Equal(0, errCode)

			bytes, err := afero.ReadFile(fs, tC.file)

			assert.NoError(err)
			assert.Equal(tC.content, string(bytes))
		})
	}
}
