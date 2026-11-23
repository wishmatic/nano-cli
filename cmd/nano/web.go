package main

import (
	"github.com/spf13/cobra"
	"github.com/wishmatic/nano-mcp/internal/cli"
)

func newWebCommand(deps cli.Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Search, fetch, map, and crawl the web",
	}

	cmd.AddCommand(
		newWebSearchCommand(deps),
		newWebScrapeCommand(deps),
		newWebMapCommand(deps),
		newWebCrawlCommand(deps),
		newWebYouTubeCommand(deps),
	)

	return cmd
}
