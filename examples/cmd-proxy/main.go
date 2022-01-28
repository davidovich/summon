package main

import (
	"embed"
	"os"

	"github.com/davidovich/summon"
)

// capture the files of the assets tree, assuming the assets directory
// is named "assets".
//go:embed assets/*
var fs embed.FS

func main() {
	os.Exit(summon.Main(os.Args, fs, summon.WithoutRunCmd()))
}
