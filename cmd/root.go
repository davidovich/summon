package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/davidovich/summon/pkg/config"
)

// CreateRoot creates the root command
func createRootCmd() *cobra.Command {
	biName := filepath.Base(os.Args[0])
	root := &cobra.Command{
		Use:   biName + " [file to summon]",
		Short: biName + " main command",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fn := args[0]
			b, err := config.Box().Open(fn)

			if err != nil {
				return err
			}

			// Write the file and print it's path
			fp := filepath.Join(config.OutputDir, fn)
			err = os.MkdirAll(config.OutputDir, os.ModePerm)
			if err != nil {
				return err
			}

			out, err := os.Create(fp)
			if err != nil {
				return err
			}
			defer out.Close()

			_, err = io.Copy(out, b)
			if err != nil {
				return err
			}

			fmt.Println(fp)
			return out.Close()
		},
	}

	//root.Flags().BoolVarP()

	root.AddCommand(newListCmd())

	return root
}

var rootCmd = createRootCmd()

// Execute is the main command entry point
func Execute() error {
	return rootCmd.Execute()
}

func init() {

}
