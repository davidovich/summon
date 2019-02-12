package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/davidovich/summon/pkg/summon"
	"github.com/gobuffalo/packr/v2"
)

func Test_createRootCmd(t *testing.T) {
	a := assert.New(t)

	rootCmd := createRootCmd(&fakeSummon{})

	a.NotNil(rootCmd)
}

type fakeSummon struct {
	*summon.Summoner
	wantErr bool
}

func (t *fakeSummon) Summon(opts ...summon.Option) (string, error) {
	if t.wantErr {
		return "", fmt.Errorf("error in Summon")
	}

	return t.Summoner.Summon()
}

func Test_mainCmd_run(t *testing.T) {
	defer testutil.ReplaceFs()()

	box := packr.New("test box", "")
	box.AddString("a", "a content")
	box.AddString("b", "b content")

	type fields struct {
		copyAll  bool
		dest     string
		filename string
		driver   *fakeSummon
	}
	tests := []struct {
		name   string
		fields fields
		out    string
	}{
		{
			name: "base",
			fields: fields{
				dest:     ".s",
				filename: "a",
				driver: &fakeSummon{
					Summoner: summon.New(box),
					wantErr:  false,
				},
			},
			out: ".s/a\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mainCmd{
				driver:   tt.fields.driver,
				copyAll:  tt.fields.copyAll,
				dest:     tt.fields.dest,
				filename: tt.fields.filename,
			}
			b := &bytes.Buffer{}
			m.out = b
			if err := m.run(); (err != nil) != tt.fields.driver.wantErr {
				t.Errorf("mainCmd.run() error = %v, wantErr %v", err, tt.fields.driver.wantErr)
			}
			assert.Equal(t, tt.out, b.String())
		})
	}
}

func TestExecute(t *testing.T) {
	type args struct {
		box *packr.Box
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Execute(tt.args.box); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
