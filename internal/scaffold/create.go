package scaffold

import (
	"embed"
	"fmt"
	"io"
	"os"

	"github.com/davidovich/summon/pkg/summon"
)

//go:embed templates/scaffold/*
var scaffoldFS embed.FS

const scaffoldParams = `{
	"ModName": "%s",
	"SummonerName": "%s",
	"go": "go"
}`

// Create will create a folder structure at destination with templates resolved
func Create(destDir, modName, summonerName string, force bool) error {
	s, err := summon.New(scaffoldFS)
	if err != nil {
		return err
	}

	empty, err := isEmptyDir(destDir)
	if !force && !empty || err != nil {
		if err == nil || !os.IsNotExist(err) {
			return fmt.Errorf("destination directory is not empty: %s", destDir)
		}
	}

	json := fmt.Sprintf(scaffoldParams, modName, summonerName)
	_, err = s.Summon(
		summon.All(true),
		summon.Filename("templates/scaffold"),
		summon.JSON(&json),
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
