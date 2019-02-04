package summon

import (
	"github.com/gobuffalo/packr/v2"

	"github.com/davidovich/summon/cmd"
)

// Main entrypoint, typically called from a data repository
func Main(args []string, box *packr.Box) int {
	err := cmd.Execute(box)

	if err != nil {
		return 1
	}

	return 0
}
