package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wishmatic/nano-mcp/internal/cli"
	"github.com/wishmatic/nano-mcp/internal/nanogpt"
)

func newWebCrawlCommand(deps cli.Deps) *cobra.Command {
	var (
		formats           []string
		proxy             string
		sitemap           string
		limit             int
		maxDiscoveryDepth int
		waitForFinish     int
		maxReturnedPages  int
	)

	cmd := &cobra.Command{
		Use:   "crawl <url>",
		Short: "Crawl a site and return the pages Firecrawl scraped",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.NanoClient(deps)
			if err != nil {
				return err
			}

			result, err := client.Firecrawl(cmd.Context(), nanogpt.FirecrawlRequest{
				Operation:         nanogpt.OperationCrawl,
				URL:               args[0],
				Formats:           formats,
				Proxy:             proxy,
				OnlyMainContent:   cli.OptionalBool(cmd, "only-main-content"),
				OnlyCleanContent:  cli.OptionalBool(cmd, "only-clean-content"),
				Sitemap:           sitemap,
				Limit:             limit,
				CrawlEntireDomain: cli.OptionalBool(cmd, "crawl-entire-domain"),
				AllowSubdomains:   cli.OptionalBool(cmd, "allow-subdomains"),
				MaxDiscoveryDepth: maxDiscoveryDepth,
				WaitForFinishSecs: waitForFinish,
				MaxReturnedPages:  maxReturnedPages,
			})
			if err != nil {
				return fmt.Errorf("web crawl: %w", err)
			}

			return cli.WriteJSON(deps.Out, result)
		},
	}

	cmd.Flags().StringArrayVar(&formats, "format", nil, "output format, such as markdown or html")
	cmd.Flags().StringVar(&proxy, "proxy", "", "Firecrawl proxy tier, basic or stealth")
	cmd.Flags().Bool("only-main-content", false, "drop navigation, footers, and other boilerplate")
	cmd.Flags().Bool("only-clean-content", false, "run NanoGPT's extra cleanup pass")
	cmd.Flags().StringVar(
		&sitemap, "sitemap", "", "include the sitemap, skip it, or use only the sitemap",
	)
	cmd.Flags().IntVar(&limit, "limit", 0, "maximum pages to crawl")
	cmd.Flags().Bool("crawl-entire-domain", false, "follow links across the whole domain")
	cmd.Flags().Bool("allow-subdomains", false, "also crawl subdomains")
	cmd.Flags().IntVar(&maxDiscoveryDepth, "max-discovery-depth", 0, "how many link hops to follow")
	cmd.Flags().IntVar(
		&waitForFinish, "wait-for-finish", 0, "seconds to wait for the crawl, 120 default",
	)
	cmd.Flags().IntVar(
		&maxReturnedPages, "max-returned-pages", 0, "pages to return after the crawl settles",
	)

	return cmd
}
