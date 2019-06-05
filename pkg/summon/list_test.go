package summon

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/gobuffalo/packr/v2"
	"github.com/stretchr/testify/assert"
)

func TestSummonerList(t *testing.T) {
	dir, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(dir)

	box := packr.New("test box", dir)
	box.AddString("a", "a content")
	box.AddString("b", "b content")

	s, _ := New(box)
	got, _ := s.List()

	assert.Equal(t, got, []string{"a", "b"}, "Summoner.List() = %v, want %v", got, []string{"a", "b"})
}

func TestSummonerListTree(t *testing.T) {
	box := packr.New("TestSummonerListTree", "nonexisting/../abc")
	box.AddString("a/b/c.txt", "")
	box.AddString("d", "")

	s, _ := New(box)

	treeList, _ := s.List(ShowTree(true))

	tree := strings.Join(treeList, "\n")

	expected := `abc
├── a
│   └── b
│       └── c.txt
└── d`

	assert.Equal(t, expected, tree)
}
