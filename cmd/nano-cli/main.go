package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wishmatic/nano-mcp/internal/cli"
)

func main() {
	if err := newRoot(cli.DefaultDeps()).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRoot(deps cli.Deps) *cobra.Command {
	deps = cli.WithDefaults(deps)

	root := &cobra.Command{
		Use:           "nano-cli",
		Short:         "Use NanoGPT from the command line",
		Long:          "Search, scrape, view, generate, and more through the NanoGPT service!",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.SetIn(deps.In)
	root.SetOut(deps.Out)
	root.SetErr(deps.Err)

	root.AddCommand(
		newLoginCommand(deps),
		newLogoutCommand(deps),
		newServeCommand(),
		newWebCommand(deps),
		newModelsCommand(deps),
		newGenCommand(deps),
	)

	return root
}
