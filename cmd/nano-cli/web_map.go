package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wishmatic/nano-mcp/internal/cli"
	"github.com/wishmatic/nano-mcp/internal/nanogpt"
)

func newWebMapCommand(deps cli.Deps) *cobra.Command {
	var (
		sitemap string
		limit   int
	)

	cmd := &cobra.Command{
		Use:   "map <url>",
		Short: "List the URLs Firecrawl can discover on a site",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.NanoClient(deps)
			if err != nil {
				return err
			}

			result, err := client.Firecrawl(cmd.Context(), nanogpt.FirecrawlRequest{
				Operation: nanogpt.OperationMap,
				URL:       args[0],
				Sitemap:   sitemap,
				Limit:     limit,
			})
			if err != nil {
				return fmt.Errorf("web map: %w", err)
			}

			return cli.WriteJSON(deps.Out, result)
		},
	}

	cmd.Flags().StringVar(
		&sitemap, "sitemap", "", "include the sitemap, skip it, or use only the sitemap",
	)
	cmd.Flags().IntVar(&limit, "limit", 0, "maximum URLs to return")

	return cmd
}
