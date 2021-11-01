package summon

import (
	"io/fs"
	"path/filepath"
	"strings"

	gotree "github.com/DiSiqueira/GoTree"
)

// List lists the content of the data tree.
func (d *Driver) List(opts ...Option) ([]string, error) {
	d.Configure(opts...)

	var list []string
	err := fs.WalkDir(d.box, d.baseDataDir, func(path string, de fs.DirEntry, err error) error {
		if path == d.baseDataDir {
			return nil
		}

		// old packr.Box List would only produce actual entries, and not intermediate
		// entries, simulate that by ignoring dir only entries.
		if !de.IsDir() {
			rel, err := filepath.Rel(d.baseDataDir, path)
			if err != nil {
				return err
			}
			list = append(list, rel)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if d.opts.tree {
		assetDir := d.baseDataDir
		rootTree := &fileTree{
			Tree:     gotree.New(assetDir),
			children: map[string]*fileTree{},
		}
		for _, p := range list {
			rootTree.addPath(p)
		}
		return []string{strings.TrimSpace(rootTree.Print())}, nil
	}

	return list, nil
}

type fileTree struct {
	gotree.Tree
	children map[string]*fileTree
}

func (tree *fileTree) addPath(path string) {
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
