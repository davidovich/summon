package cmd

import (
	"fmt"
	"strings"

	"github.com/davidovich/summon/pkg/config"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	lcmd := &cobra.Command{
		Use:   "list",
		Short: "list all summonables",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(strings.Join(config.Box().List(), "\n"))

			return nil
		},
	}
	return lcmd
}
