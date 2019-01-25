package summon

import (
	"os"

	"github.com/gobuffalo/packr/v2"
)

// Main entrypoint
func Main(args []string) int {
	root := os.Args[1]

	packr.New("Main Box", root)
	return 0
}

// AddAssetRoot adds an asset root dir that will be bundled
func AddAssetRoot(root string) {

}
