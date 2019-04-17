package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
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
	json     string
	jsonFile string
	raw      bool
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
			if main.json != "" && main.jsonFile != "" {
				return fmt.Errorf("--json and --json-file are mutually exclusive")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			main.out = cmd.OutOrStdout()
			if !main.copyAll {
				filename := args[0]
				main.filename = filename
			}
			if main.jsonFile != "" {
				var j []byte
				var err error
				if main.jsonFile == "-" {
					j, err = ioutil.ReadAll(os.Stdin)
				} else {
					j, err = ioutil.ReadFile(main.jsonFile)
				}
				if err != nil {
					return err
				}

				main.json = string(j)
			}

			return main.run()
		},
	}

	rootCmd.Flags().StringVar(&main.json, "json", "", "json to use to render template")
	rootCmd.Flags().StringVar(&main.jsonFile, "json-file", "", "json file to use to render template, with '-' for stdin")
	rootCmd.Flags().BoolVarP(&main.copyAll, "all", "a", false, "restitute all data")
	rootCmd.Flags().BoolVar(&main.raw, "raw", false, "output without any template rendering")
	rootCmd.Flags().StringVarP(&main.dest, "out", "o", config.OutputDir, "destination directory, or '-' for stdout")

	rootCmd.AddCommand(newListCmd(driver))
	rootCmd.AddCommand(newRunCmd(driver))

	return rootCmd
}

func (m *mainCmd) run() error {
	m.driver.Configure(
		summon.All(m.copyAll),
		summon.Dest(m.dest),
		summon.Filename(m.filename),
		summon.JSON(m.json),
		summon.Raw(m.raw),
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
	s, err := summon.New(box)
	if err != nil {
		return err
	}
	rootCmd := createRootCmd(s)
	return rootCmd.Execute()
}
