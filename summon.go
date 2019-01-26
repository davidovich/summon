package summon

import (
	"fmt"

	"github.com/gobuffalo/packr/v2"
)

// Main entrypoint
func Main(args []string, box *packr.Box) int {
	s, err := box.FindString(args[1])

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 1
	}

	fmt.Println(s)
	return 0
}
