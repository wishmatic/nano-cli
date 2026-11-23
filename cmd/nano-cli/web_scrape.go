package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wishmatic/nano-mcp/internal/cli"
	"github.com/wishmatic/nano-mcp/internal/nanogpt"
)

func newWebScrapeCommand(deps cli.Deps) *cobra.Command {
	var (
		firecrawl bool
		provider  string
		stealth   bool
		formats   []string
		proxy     string
	)

	cmd := &cobra.Command{
		Use:   "scrape <url...>",
		Short: "Fetch the readable text of web pages",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if firecrawl {
				return runFirecrawlScrape(deps, cmd, args, formats, proxy)
			}

			return runScrape(deps, cmd, args, provider, stealth)
		},
	}

	cmd.Flags().BoolVar(&firecrawl, "firecrawl", false, "scrape one page through Firecrawl")
	cmd.Flags().StringVar(
		&provider, "provider", "", "auto or linkup; auto routes between the crawler and Linkup",
	)
	cmd.Flags().BoolVar(
		&stealth, "stealth", false, "route through a stealth proxy at 5x the per-URL charge",
	)
	cmd.Flags().StringArrayVar(
		&formats, "format", nil, "Firecrawl output format, such as markdown or html",
	)
	cmd.Flags().StringVar(&proxy, "proxy", "", "Firecrawl proxy tier, basic or stealth")
	cmd.Flags().Bool("only-main-content", false, "drop navigation, footers, and other boilerplate")
	cmd.Flags().Bool("only-clean-content", false, "run NanoGPT's extra cleanup pass")

	return cmd
}

func runScrape(
	deps cli.Deps, cmd *cobra.Command, urls []string, provider string, stealth bool,
) error {
	if len(urls) > nanogpt.MaxScrapeURLs {
		return fmt.Errorf("web scrape takes 1 to %d URLs, got %d", nanogpt.MaxScrapeURLs, len(urls))
	}

	client, err := cli.NanoClient(deps)
	if err != nil {
		return err
	}

	response, err := client.Scrape(cmd.Context(), nanogpt.ScrapeRequest{
		URLs:        urls,
		Provider:    provider,
		StealthMode: stealth,
	})
	if err != nil {
		return fmt.Errorf("web scrape: %w", err)
	}

	return printScrape(deps, response)
}

func runFirecrawlScrape(
	deps cli.Deps, cmd *cobra.Command, urls []string, formats []string, proxy string,
) error {
	if len(urls) != 1 {
		return fmt.Errorf("firecrawl scrape takes exactly one URL, got %d", len(urls))
	}

	client, err := cli.NanoClient(deps)
	if err != nil {
		return err
	}

	result, err := client.Firecrawl(cmd.Context(), nanogpt.FirecrawlRequest{
		Operation:        nanogpt.OperationScrape,
		URL:              urls[0],
		Formats:          formats,
		Proxy:            proxy,
		OnlyMainContent:  cli.OptionalBool(cmd, "only-main-content"),
		OnlyCleanContent: cli.OptionalBool(cmd, "only-clean-content"),
	})
	if err != nil {
		return fmt.Errorf("web scrape --firecrawl: %w", err)
	}

	return cli.WriteJSON(deps.Out, result)
}

func printScrape(deps cli.Deps, response *nanogpt.ScrapeResponse) error {
	for index, result := range response.Results {
		if index > 0 {
			fmt.Fprintln(deps.Out)
		}

		fmt.Fprintf(deps.Out, "URL: %s\n", result.URL)
		fmt.Fprintf(deps.Out, "Success: %t\n", result.Success)

		if result.Title != "" {
			fmt.Fprintf(deps.Out, "Title: %s\n", result.Title)
		}

		body := result.Markdown
		label := "Markdown"
		if strings.TrimSpace(body) == "" {
			body = result.Content
			label = "Content"
		}

		if strings.TrimSpace(body) != "" {
			fmt.Fprintf(deps.Out, "%s:\n%s\n", label, body)
		}

		if result.Error != "" {
			fmt.Fprintf(deps.Out, "Error: %s\n", result.Error)
		}
	}

	fmt.Fprintln(deps.Out, "Summary:")

	return cli.WriteJSON(deps.Out, response.Summary)
}
