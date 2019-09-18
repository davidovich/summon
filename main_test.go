// summon_test is an example entry-point file for a summon asset repository.
// You should replace `package summon_test` with `package main` in your data
// repo implementation.
package summon_test // replace with package main

import (
	"os"

	"github.com/davidovich/summon"
	"github.com/gobuffalo/packr/v2"
)

var exit = os.Exit

// Here is what the bootstrapped summon.go will look like:
//
func Example() { // Example() should be replaced by main()

	// box captures the files of the assets tree
	box := packr.New("Summon Box", "../assets")

	// relinquish control to the summon library
	exit(summon.Main(os.Args, box))
}
