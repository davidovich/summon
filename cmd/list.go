package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/davidovich/summon/pkg/summon"
)

type listCmdOpts struct {
	driver   summon.ConfigurableLister
	tree     bool
	asOption bool
	out      io.Writer
	cmd      *cobra.Command
}

func newListCmd(asOption bool, root *cobra.Command, driver summon.ConfigurableLister, main *mainCmd) {
	listCmd := &listCmdOpts{
		driver: driver,
	}
	main.listOptions = listCmd

	if asOption {
		root.Flags().BoolVarP(&listCmd.asOption, "ls", "", false, "list embedded summonables")
	} else {
		lcmd := &cobra.Command{
			Use:   "ls",
			Short: "List all summonables",
			RunE: func(cmd *cobra.Command, args []string) error {
				listCmd.out = cmd.OutOrStdout()
				return listCmd.run()
			},
		}
		listCmd.cmd = lcmd
		root.AddCommand(lcmd)
		root = lcmd
	}

	listCmd.out = root.OutOrStdout()
	root.Flags().BoolVar(&listCmd.tree, "tree", false, "Print pretty tree of data")
}

func (l *listCmdOpts) run() error {
	l.driver.Configure(
		summon.ShowTree(l.tree),
	)

	list, err := l.driver.List()
	if err != nil {
		return err
	}

	fmt.Fprintln(l.out, strings.Join(list, "\n"))
	return nil
}
