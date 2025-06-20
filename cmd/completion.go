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
		Use:   "completion [bash|zsh]",
		Short: "Generate completion script",
		Long: `To load completion, run

. <(summon completion bash)

or

. <(summon completion zsh)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(summon completion bash)

# ~/.zshrc
. <(summon completion zsh)
`,
		ValidArgs: []string{"bash", "zsh"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			cOpts.out = cmd.OutOrStdout()
			cOpts.cmd = cmd
			return cOpts.run(args[0])
		},
	}

	return c
}

func (c completionCmdOpts) run(shell string) error {
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

	switch shell {
	case "bash":
		return c.cmd.Root().GenBashCompletionV2(c.out, true)
	case "zsh":
		return c.cmd.Root().GenZshCompletion(c.out)
	}

	return nil
}
