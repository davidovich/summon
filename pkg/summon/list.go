package summon

import (
	"path/filepath"
	"strings"

	gotree "github.com/DiSiqueira/GoTree"
)

// List list the content of the data tree.
func (s *Summoner) List(opts ...Option) ([]string, error) {
	s.Configure(opts...)

	list := s.box.List()

	if s.opts.tree {
		_, assetDir := filepath.Split(s.box.Path)
		rootTree := &fileTree{
			Tree:     gotree.New(assetDir),
			children: map[string]*fileTree{},
		}
		for _, path := range list {
			rootTree.AddPath(path)
		}
		return []string{strings.TrimSpace(rootTree.Print())}, nil
	}

	return list, nil
}

type fileTree struct {
	gotree.Tree
	children map[string]*fileTree
}

func (tree *fileTree) AddPath(path string) {
	// first get path components
	comp := strings.Split(path, "/")

	for i, d := range comp {
		if i == len(comp)-1 { // file
			tree.Add(d)
			break
		}
		child, ok := tree.children[d]
		if !ok {
			child = &fileTree{Tree: tree.Add(d), children: map[string]*fileTree{}}
			tree.children[d] = child
		}
		tree = child
	}
}
