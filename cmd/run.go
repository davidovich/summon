package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/davidovich/summon/pkg/summon"
)

func newRunCmd(runCmdEnabled bool, root *cobra.Command, driver summon.ConfigurableRunner, main *mainCmd) (*cobra.Command, error) {
	osArgs := os.Args
	if main.osArgs != nil {
		osArgs = *main.osArgs
	}
	// calculate the extra args to pass to the referenced executable
	// this is due to a limitation in spf13/cobra which eats
	// all unknown args or flags making it hard to wrap other commands.
	// We are lucky, we know the prefix order of params,
	// extract args after the run command [summon run handle]
	// see https://github.com/spf13/pflag/pull/160
	// https://github.com/spf13/cobra/issues/739
	// and https://github.com/spf13/pflag/pull/199
	firstUnknownArgPos := 2
	if runCmdEnabled {
		firstUnknownArgPos = 3
	}
	if strings.Contains(strings.Join(osArgs, " "), "-test.run") {
		firstUnknownArgPos++
	}
	if firstUnknownArgPos > len(osArgs) {
		firstUnknownArgPos = len(osArgs)
	}
	userArgs := osArgs[firstUnknownArgPos:]

	// read config for exec section
	driver.Configure(
		summon.Args(userArgs...),
		// TODO: JSON is set too soon because the parsing of flags might
		// change the captured value.
		summon.JSON(main.json),
	)

	return driver.ConstructCommandTree(root, runCmdEnabled)
}
