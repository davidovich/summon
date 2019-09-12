package scaffold

import (
	"fmt"
	"io"
	"os"

	"github.com/davidovich/summon/pkg/summon"
	"github.com/gobuffalo/packr/v2"
)

const scaffoldParams = `{
	"ModName": "%s",
	"SummonerName": "%s",
	"go": ".go"
}`

// Create will create a folder structure at destination with templates resolved
func Create(destDir, modName, summonerName string, force bool) error {
	box := packr.New("Summon scaffold template", "../templates/scaffold")

	s, _ := summon.New(box)

	empty, err := isEmptyDir(destDir)
	if !force && !empty || err != nil {
		if err == nil || !os.IsNotExist(err) {
			return fmt.Errorf("destination directory is not empty")
		}
	}

	_, err = s.Summon(
		summon.All(true),
		summon.JSON(fmt.Sprintf(scaffoldParams, modName, summonerName)),
		summon.Dest(destDir),
	)

	return err
}

// https://stackoverflow.com/a/30708914/28275
func isEmptyDir(dir string) (bool, error) {
	f, err := summon.GetFs().Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
