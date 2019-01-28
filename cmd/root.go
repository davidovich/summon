package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/davidovich/summon/pkg/config"
)

// CreateRoot creates the root command
func createRoot() *cobra.Command {
	biName := filepath.Base(os.Args[0])
	return &cobra.Command{
		Use:   biName + " [file to summon]",
		Short: biName + " main command",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := config.Box().FindString(args[0])

			if err != nil {
				return err
			}

			fmt.Println(s)
			return nil
		},
	}
}

var rootCmd = createRoot()

// Execute is the main command entry point
func Execute() error {
	return rootCmd.Execute()
}

func init() {

}
