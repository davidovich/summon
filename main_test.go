// This is an example entry-point file for a summon asset repository.
// This file can be bootrapped with:
//   go run github.com/davidovich/summon/scaffold init [assets module name]
package summon_test

import (
	"embed"
	"os"

	"github.com/davidovich/summon"
)

// fs captures the files of the assets tree
//go:embed assets/*
var fs embed.FS

var exit = os.Exit

// Here is the main() entry-point in summon.go
func Example() {
	// relinquish control to the summon library
	exit(summon.Main(os.Args, fs))
}
