package summon

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/gobuffalo/packr/v2"
	"github.com/stretchr/testify/assert"
)

func TestSummoner_List(t *testing.T) {
	defer testutil.ReplaceFs()()

	dir, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(dir)

	box := packr.New("test box", dir)
	box.AddString("a", "a content")
	box.AddString("b", "b content")

	s, _ := New(box)
	got, _ := s.List()

	assert.Equal(t, got, []string{"a", "b"}, "Summoner.List() = %v, want %v", got, []string{"a", "b"})
}
