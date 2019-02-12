package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gobuffalo/packr/v2"
	"github.com/spf13/cobra"

	"github.com/davidovich/summon/pkg/config"
	"github.com/davidovich/summon/pkg/summon"
)

type mainCmd struct {
	copyAll  bool
	dest     string
	driver   summon.Interface
	filename string
	out      io.Writer
}

// CreateRoot creates the root command
func createRootCmd(driver summon.Interface) *cobra.Command {
	cmdName := filepath.Base(os.Args[0])

	main := &mainCmd{
		driver: driver,
	}

	rootCmd := &cobra.Command{
		Use:   cmdName + " [file to summon]",
		Short: cmdName + " main command",
		Args: func(cmd *cobra.Command, args []string) error {
			if main.copyAll {
				return nil
			}
			if len(args) < 1 {
				return fmt.Errorf("requires one file to summon, received %d", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			main.out = cmd.OutOrStdout()
			if !main.copyAll {
				filename := args[0]
				main.filename = filename
			}

			return main.run()
		},
	}

	rootCmd.Flags().BoolVarP(&main.copyAll, "all", "a", false, "restitute all data")
	rootCmd.Flags().StringVar(&main.dest, "to", config.OutputDir, "destination directory")

	rootCmd.AddCommand(newListCmd(driver))
	rootCmd.AddCommand(newRunCmd(driver))

	return rootCmd
}

func (m *mainCmd) run() error {
	m.driver.Configure(
		summon.All(m.copyAll),
		summon.Dest(m.dest),
		summon.Filename(m.filename),
	)

	resultFilepath, err := m.driver.Summon()
	if err != nil {
		return err
	}
	fmt.Fprintln(m.out, resultFilepath)
	return nil
}

// Execute is the main command entry point
func Execute(box *packr.Box) error {
	rootCmd := createRootCmd(summon.New(box))
	return rootCmd.Execute()
}
