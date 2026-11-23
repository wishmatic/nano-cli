package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wishmatic/nano-mcp/internal/cli"
)

func newGenCommand(deps cli.Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gen",
		Aliases: []string{"generate"},
		Short:   "Generate images and videos",
	}

	cmd.AddCommand(
		newGenImageCommand(deps),
		newGenVideoCommand(deps),
	)

	return cmd
}

func promptFromArgs(args []string) (string, error) {
	prompt := strings.TrimSpace(strings.Join(args, " "))
	if prompt == "" {
		return "", fmt.Errorf("prompt is empty")
	}

	return prompt, nil
}
