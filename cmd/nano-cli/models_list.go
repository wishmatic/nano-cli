package main

import (
	"fmt"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/wishmatic/nano-mcp/internal/cli"
	"github.com/wishmatic/nano-mcp/internal/nanogpt"
)

const (
	defaultModelLimit = 25
	maxModelLimit     = 100
)

func newModelsListCommand(deps cli.Deps) *cobra.Command {
	var (
		query string
		limit int
	)

	cmd := &cobra.Command{
		Use:   "list [image|video]",
		Short: "List the image or video models NanoGPT serves",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kinds, err := modelKinds(args)
			if err != nil {
				return err
			}

			shownLimit, err := modelLimit(limit)
			if err != nil {
				return err
			}

			client, err := cli.NanoClient(deps)
			if err != nil {
				return err
			}

			for _, kind := range kinds {
				all, err := client.Models(cmd.Context(), kind)
				if err != nil {
					return fmt.Errorf("models list %s: %w", kind, err)
				}

				matched := filterModels(all, query)
				shown := matched
				if len(shown) > shownLimit {
					shown = shown[:shownLimit]
				}

				if err := printModels(deps, kind, all, matched, shown); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "filter by id, name, description, owner, or tags")
	cmd.Flags().IntVar(&limit, "limit", 0, "models to show per kind, 25 default and 100 at most")

	return cmd
}

func modelKinds(args []string) ([]string, error) {
	if len(args) == 0 {
		return nanogpt.ModelKinds, nil
	}

	kind := strings.ToLower(strings.TrimSpace(args[0]))
	if !slices.Contains(nanogpt.ModelKinds, kind) {
		return nil, fmt.Errorf("model kind %q is not one of %v", kind, nanogpt.ModelKinds)
	}

	return []string{kind}, nil
}

func modelLimit(limit int) (int, error) {
	switch {
	case limit == 0:
		return defaultModelLimit, nil
	case limit < 0 || limit > maxModelLimit:
		return 0, fmt.Errorf("limit must be 1 to %d, got %d", maxModelLimit, limit)
	default:
		return limit, nil
	}
}

func filterModels(models []nanogpt.Model, query string) []nanogpt.Model {
	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return models
	}

	matched := make([]nanogpt.Model, 0, len(models))

	for _, model := range models {
		if modelMatches(model, needle) {
			matched = append(matched, model)
		}
	}

	return matched
}

func modelMatches(model nanogpt.Model, needle string) bool {
	fields := append([]string{model.ID, model.Name, model.Description, model.OwnedBy}, model.Tags...)

	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), needle) {
			return true
		}
	}

	return false
}

func printModels(
	deps cli.Deps, kind string, all, matched, shown []nanogpt.Model,
) error {
	fmt.Fprintf(deps.Out, "%s MODELS\n", strings.ToUpper(kind))

	writer := tabwriter.NewWriter(deps.Out, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(writer, "ID\tNAME\tOWNER\tDESCRIPTION"); err != nil {
		return err
	}

	for _, model := range shown {
		if _, err := fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			tableField(model.ID), tableField(model.Name), tableField(model.OwnedBy),
			tableField(model.Description),
		); err != nil {
			return err
		}
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	fmt.Fprintf(deps.Out, "%d of %d models shown (%d matched)\n\n", len(shown), len(all), len(matched))

	return nil
}

func tableField(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if value == "" {
		return "-"
	}

	return value
}
