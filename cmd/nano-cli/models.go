package main

import (
	"github.com/spf13/cobra"
	"github.com/wishmatic/nano-mcp/internal/cli"
)

func newModelsCommand(deps cli.Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "List NanoGPT models and read the account balance",
	}

	cmd.AddCommand(
		newModelsListCommand(deps),
		newModelsBalanceCommand(deps),
	)

	return cmd
}
