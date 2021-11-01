package main

import (
	"os"

	"github.com/davidovich/summon"
	"github.com/gobuffalo/packr/v2"
)

func main() {
	// Box captures the files of the assets tree
	box := packr.New("Summon Box", "../assets")
	os.Exit(summon.Main(os.Args, box))
}
