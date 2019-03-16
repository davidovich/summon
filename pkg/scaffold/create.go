package scaffold

import (
	"github.com/davidovich/summon/pkg/summon"
	"github.com/gobuffalo/packr/v2"
)

type scaffoldParams struct {
	ModName      string
	SummonerName string
}

// Create will create a folder structure at destination with templates resolved
func Create(destDir string, force bool) error {
	box := packr.New("Summon scaffold template", "../../templates/scaffold")

	s, _ := summon.New(box)

	_, err := s.Summon()
	return err
}
