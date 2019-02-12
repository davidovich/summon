package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/davidovich/summon/pkg/summon"
)

type listCmdOpts struct {
	driver summon.Interface
	tree   bool
	out    io.Writer
}

func newListCmd(driver summon.Interface) *cobra.Command {
	listCmd := &listCmdOpts{
		driver: driver,
	}
	lcmd := &cobra.Command{
		Use:   "list",
		Short: "list all summonables",
		RunE: func(cmd *cobra.Command, args []string) error {
			listCmd.out = cmd.OutOrStdout()
			return listCmd.run()
		},
	}

	lcmd.Flags().BoolVar(&listCmd.tree, "tree", false, "Print pretty tree of data")

	return lcmd
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
