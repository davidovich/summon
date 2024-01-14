// Package cmd defines the main command line interface entry-points.
//
// See summon -h:
//
//	summon main command
//
//	Usage:
//	  summon [file to summon] [flags]
//	  summon [command]
//
//	Available Commands:
//	  completion  Output bash completion script
//	  help        Help about any command
//	  ls          List all summonables
//	  run         Launch executable from summonables
//
//	Flags:
//	  -a, --all                restitute all data
//	  -h, --help               help for summon
//	      --json string        json to use to render template
//	      --json-file string   json file to use to render template, with '-' for stdin
//	  -o, --out string         destination directory, or '-' for stdout (default ".summoned")
//	      --raw                output without any template rendering
//	  -v, --version            output data version info and exit
//
//	Use "summon [command] --help" for more information about a command.
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"github.com/davidovich/summon/pkg/summon"
)

type mainCmd struct {
	copyAll     bool
	dest        string
	driver      *summon.Driver
	filename    string
	raw         bool
	debug       bool
	out         io.Writer
	osArgs      *[]string
	listOptions *listCmdOpts
	cmd         *cobra.Command
}

// CreateRootCmd creates the root command
func CreateRootCmd(driver *summon.Driver, args []string, options summon.MainOptions) (*cobra.Command, error) {
	exeName := filepath.Base(args[0])
	var showVersion bool

	main := &mainCmd{
		driver: driver,
		osArgs: &args,
	}

	cmdHint := " [file to summon]"
	if options.WithoutRunSubcmd {
		cmdHint = " [handle | file to summon]"
	}
	rootCmd := &cobra.Command{
		Use:              exeName + cmdHint,
		Short:            exeName + " main command",
		TraverseChildren: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if main.copyAll || showVersion || main.listOptions.asOption || main.listOptions.tree {
				return nil
			}
			if len(args) < 1 {
				return fmt.Errorf("requires one file to summon, received %d", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				v, ok := makeVersion()
				if !ok {
					return fmt.Errorf("missing build info")
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				enc.Encode(v)
				return nil
			}
			if len(args) != 0 {
				main.filename = args[0]
			}
			return main.run()
		},
	}

	main.cmd = rootCmd

	// rootCmd.PersistentFlags().BoolVarP(&main.debug, "debug", "d", false, "print debug info on stderr")
	rootCmd.Flags().BoolVarP(&main.copyAll, "all", "a", false, "restitute all data")
	rootCmd.Flags().BoolVar(&main.raw, "raw", false, "output without any template rendering")
	rootCmd.Flags().StringVarP(&main.dest, "out", "o", driver.OutputDir(), "destination directory, or '-' for stdout")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "output data version info and exit")

	// we have a --help flag, hide the help sub-command
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	// add ls cmd or --ls flag
	newListCmd(options.WithoutRunSubcmd, rootCmd, driver, main)

	if !driver.HideAssetsInHelp() {
		// configure summonables completion
		list, _ := driver.List()
		for _, summonable := range list {
			rootCmd.AddCommand(&cobra.Command{
				Use:   summonable,
				Short: "summon " + summonable + " file to " + main.dest + "/ dir",
				RunE: func(cmd *cobra.Command, args []string) error {
					main.filename = cmd.Use
					return main.run()
				}})
		}
	} else {
		rootCmd.ValidArgsFunction = func(cmd *cobra.Command, cobraArgs []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			var candidates []string
			list, _ := driver.List()
			for _, summonable := range list {
				if strings.HasPrefix(summonable, toComplete) {
					candidates = append(candidates, summonable)
				}
			}
			return candidates, cobra.ShellCompDirectiveNoFileComp
		}
	}

	// add run cmd, or root subcommands
	runRoot, err := newRunCmd(!options.WithoutRunSubcmd, rootCmd, driver, main)
	if err != nil {
		return nil, err
	}

	// add completion
	rootCmd.AddCommand(newCompletionCmd(driver))

	// ask driver to register its flags
	driver.RegisterFlags(runRoot)

	driver.SetupRunArgs(runRoot)

	return rootCmd, nil
}

func (m *mainCmd) run() error {
	// tree implies ls
	if m.listOptions.asOption || m.listOptions.tree {
		err := m.listOptions.run()
		if err != nil {
			return err
		}
		return nil
	}

	if m.out == nil {
		m.out = m.cmd.OutOrStdout()
	}

	err := m.driver.Configure(
		summon.All(m.copyAll),
		summon.Dest(m.dest),
		summon.Filename(m.filename),
		summon.Raw(m.raw),
	)
	if err != nil {
		return err
	}

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
