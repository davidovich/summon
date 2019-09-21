// This is an example entry-point file for a summon asset repository.
// This file can be bootrapped with:
//   go run github.com/davidovich/summon/scaffold init [assets module name]
package summon_test

import (
	"os"

	"github.com/davidovich/summon"
	"github.com/gobuffalo/packr/v2"
)

var exit = os.Exit

// Here is the main() entry-point in summon.go
func Example() {

	// box captures the files of the assets tree
	box := packr.New("Summon Box", "../assets")

	// relinquish control to the summon library
	exit(summon.Main(os.Args, box))
}
