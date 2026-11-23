package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wishmatic/nano-mcp/internal/cli"
	"github.com/wishmatic/nano-mcp/internal/nanogpt"
)

func newWebSearchCommand(deps cli.Deps) *cobra.Command {
	var (
		provider   string
		depth      string
		outputType string
	)

	cmd := &cobra.Command{
		Use:   "search <query...>",
		Short: "Search the web through NanoGPT's providers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(strings.Join(args, " "))
			if query == "" {
				return fmt.Errorf("search query is empty")
			}

			client, err := cli.NanoClient(deps)
			if err != nil {
				return err
			}

			response, err := client.Search(cmd.Context(), nanogpt.SearchRequest{
				Query:      query,
				Provider:   provider,
				Depth:      depth,
				OutputType: outputType,
			})
			if err != nil {
				return fmt.Errorf("web search: %w", err)
			}

			return printSearch(deps, response)
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "provider to use, such as linkup or exa")
	cmd.Flags().StringVar(&depth, "depth", "", "search depth, standard unless deepened")
	cmd.Flags().StringVar(
		&outputType, "output-type", "", "searchResults, sourcedAnswer, or structured",
	)

	return cmd
}

func printSearch(deps cli.Deps, response *nanogpt.SearchResponse) error {
	meta := response.Metadata

	fmt.Fprintf(deps.Out, "Query: %s\n", meta.Query)
	fmt.Fprintf(deps.Out, "Provider: %s\n", meta.Provider)
	fmt.Fprintf(deps.Out, "Operation: %s\n", meta.Operation)
	fmt.Fprintf(deps.Out, "Depth: %s\n", meta.Depth)
	fmt.Fprintf(deps.Out, "Output type: %s\n", meta.OutputType)
	fmt.Fprintf(deps.Out, "Timestamp: %s\n", meta.Timestamp)
	fmt.Fprintf(deps.Out, "Cost: %s\n", cli.FormatUSD(meta.CostUSD))
	fmt.Fprintln(deps.Out, "Data:")

	return cli.WriteJSON(deps.Out, response.Data)
}
