package summon

import (
	"github.com/gobuffalo/packr/v2"

	"github.com/davidovich/summon/cmd"
	"github.com/davidovich/summon/pkg/config"
)

// Main entrypoint, typically called from a data repository
func Main(args []string, box *packr.Box) int {

	config.SetBox(box)

	err := cmd.Execute()

	if err != nil {
		return 1
	}

	return 0
}
