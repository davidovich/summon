// Package cmd defines the main command line interface entry-points.
//
// See summon -h:
//  summon main command
//
//  Usage:
//    summon [file to summon] [flags]
//    summon [command]
//
//  Available Commands:
//    completion  Output bash completion script
//    help        Help about any command
//    ls          List all summonables
//    run         Launch executable from summonables
//
//  Flags:
//    -a, --all                restitute all data
//    -h, --help               help for summon
//        --json string        json to use to render template
//        --json-file string   json file to use to render template, with '-' for stdin
//    -o, --out string         destination directory, or '-' for stdout (default ".summoned")
//        --raw                output without any template rendering
//    -v, --version            output data version info and exit
//
//  Use "summon [command] --help" for more information about a command.
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/davidovich/summon/pkg/summon"
)

type mainCmd struct {
	copyAll     bool
	dest        string
	driver      *summon.Driver
	filename    string
	json        string
	jsonFile    string
	raw         bool
	debug       bool
	out         io.Writer
	osArgs      *[]string
	listOptions *listCmdOpts
}

// CreateRootCmd creates the root command
func CreateRootCmd(driver *summon.Driver, args []string, options summon.MainOptions) (*cobra.Command, error) {
	cmdName := filepath.Base(args[0])
	var showVersion bool

	main := &mainCmd{
		driver: driver,
	}

	rootCmd := &cobra.Command{
		Use:   cmdName + " [file to summon]",
		Short: cmdName + " main command",
		Args: func(cmd *cobra.Command, args []string) error {
			if main.copyAll || showVersion || main.listOptions.asOption {
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
			if showVersion {
				v, ok := makeVersion()
				if !ok {
					return fmt.Errorf("Missing build info")
				}
				enc := json.NewEncoder(main.out)
				enc.SetIndent("", "  ")
				enc.Encode(v)
				return nil
			}
			if !main.copyAll && !main.listOptions.asOption {
				filename := args[0]
				main.filename = filename
			}
			if main.jsonFile != "" {
				var j []byte
				var err error
				if main.jsonFile == "-" {
					j, err = ioutil.ReadAll(cmd.InOrStdin())
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

	// TODO These should probably be simple flags (not persistent)
	rootCmd.PersistentFlags().StringVar(&main.json, "json", "", "json to use to render template")
	rootCmd.PersistentFlags().StringVar(&main.jsonFile, "json-file", "", "json file to use to render template, with '-' for stdin")
	rootCmd.PersistentFlags().BoolVarP(&main.debug, "debug", "d", false, "print debug info on stderr")
	rootCmd.Flags().BoolVarP(&main.copyAll, "all", "a", false, "restitute all data")
	rootCmd.Flags().BoolVar(&main.raw, "raw", false, "output without any template rendering")
	rootCmd.Flags().StringVarP(&main.dest, "out", "o", driver.OutputDir(), "destination directory, or '-' for stdout")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "output data version info and exit")

	// add ls cmd or --ls flag
	main.listOptions = newListCmd(options.WithoutRunSubcmd, rootCmd, driver)

	// add run cmd, or root subcommands
	_, err := newRunCmd(!options.WithoutRunSubcmd, rootCmd, driver, main)

	// add completion
	rootCmd.AddCommand(newCompletionCmd(driver))

	return rootCmd, err
}

func (m *mainCmd) run() error {
	if m.listOptions.asOption {
		err := m.listOptions.run()
		if err != nil {
			return err
		}
		return nil
	}

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

type versionDesc struct {
	Exe     string `json:"exe,omitempty"`
	Mod     string `json:"mod"`
	Version string `json:"version"`
}
type versionInfo struct {
	Assets versionDesc `json:"assets"`
	Lib    versionDesc `json:"lib"`
}

var buildInfo = debug.ReadBuildInfo

func makeVersion() (v versionInfo, ok bool) {
	bi, ok := buildInfo()
	if !ok {
		return v, false
	}
	v = versionInfo{
		Lib: versionDesc{},
		Assets: versionDesc{
			Mod:     bi.Main.Path,
			Version: bi.Main.Version,
			Exe:     os.Args[0],
		},
	}
	for _, d := range bi.Deps {
		if d.Path == "github.com/davidovich/summon" {
			v.Lib.Mod = d.Path
			v.Lib.Version = d.Version
			break
		}
	}
	return v, true
}
