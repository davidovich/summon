package cmd

import (
	"io"

	"github.com/davidovich/summon/pkg/summon"
	"github.com/spf13/cobra"
)

type completionCmdOpts struct {
	driver summon.Lister
	cmd    *cobra.Command
	out    io.Writer
}

func newCompletionCmd(driver summon.Lister) *cobra.Command {
	cOpts := completionCmdOpts{
		driver: driver,
	}

	c := &cobra.Command{
		Use:   "completion",
		Short: "Output bash completion script",
		Long: `To load completion run

. <(summon completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(summon completion)
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cOpts.out = cmd.OutOrStdout()
			cOpts.cmd = cmd
			return cOpts.run()
		},
	}

	return c
}

func (c completionCmdOpts) run() error {
	list, err := c.driver.List()
	if err != nil {
		return err
	}

	// make fake command structure to populate command completions for assets
	for _, f := range list {
		c.cmd.Root().AddCommand(&cobra.Command{
			Use: f,
			Run: func(cmd *cobra.Command, args []string) {},
		})
	}

	return c.cmd.Root().GenBashCompletion(c.out)
}
