package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/gobuffalo/packr/v2"
	"github.com/spf13/cobra"

	"github.com/davidovich/summon/pkg/summon"
)

type listCmdOpts struct {
	box    *packr.Box
	driver summon.Interface
	tree   bool
	out    io.Writer
}

func newListCmd(box *packr.Box, driver summon.Interface) *cobra.Command {
	listCmd := &listCmdOpts{
		box:    box,
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
	if l.driver == nil {
		l.driver = summon.New(
			l.box,
			summon.ShowTree(l.tree),
		)
	}

	list, err := l.driver.List()
	if err != nil {
		return err
	}

	fmt.Fprintln(l.out, strings.Join(list, "\n"))
	return nil
}
