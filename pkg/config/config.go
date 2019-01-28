package config

import (
	"github.com/gobuffalo/packr/v2"
)

// Box is the main box of the program
var mainBox *packr.Box

// SetBox sets the main data box
func SetBox(box *packr.Box) {
	mainBox = box
}

// Box returns the main data box
func Box() *packr.Box {
	return mainBox
}
