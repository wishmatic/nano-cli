package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wishmatic/nano-mcp/internal/cli"
)

func newModelsBalanceCommand(deps cli.Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "balance",
		Short: "Read the NanoGPT account balance",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cli.NanoClient(deps)
			if err != nil {
				return err
			}

			balance, err := client.CheckBalance(cmd.Context())
			if err != nil {
				return fmt.Errorf("models balance: %w", err)
			}

			fmt.Fprintf(deps.Out, "USD: %s\n", cli.FormatUSD(balance.USD))
			fmt.Fprintf(deps.Out, "Nano: %s\n", cli.FormatAmount(balance.Nano))

			if balance.DepositAddress != "" {
				fmt.Fprintf(deps.Out, "Deposit address: %s\n", balance.DepositAddress)
			}

			return nil
		},
	}
}
