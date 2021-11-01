package summon

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestSummonerList(t *testing.T) {
	testFs := fstest.MapFS{}

	testFs["dir/a"] = &fstest.MapFile{Data: []byte("a content")}
	testFs["dir/b"] = &fstest.MapFile{Data: []byte("b content")}

	s, _ := New(testFs)
	got, _ := s.List()

	assert.Equal(t, []string{"a", "b"}, got, "Summoner.List() = %v, want %v", got, []string{"a", "b"})
}

func TestSummonerListTree(t *testing.T) {
	testFs := fstest.MapFS{}
	testFs["abc/a/b/c.txt"] = &fstest.MapFile{}
	testFs["abc/d"] = &fstest.MapFile{}

	s, _ := New(testFs)

	treeList, _ := s.List(ShowTree(true))

	tree := strings.Join(treeList, "\n")

	expected := `abc
├── a
│   └── b
│       └── c.txt
└── d`

	assert.Equal(t, expected, tree)
}
