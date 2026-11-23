package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wishmatic/nano-mcp/internal/cli"
	"github.com/wishmatic/nano-mcp/internal/nanogpt"
)

func newWebYouTubeCommand(deps cli.Deps) *cobra.Command {
	return &cobra.Command{
		Use:     "youtube <url...>",
		Aliases: []string{"yt"},
		Short:   "Fetch the transcripts of YouTube videos",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > nanogpt.MaxTranscribeURLs {
				return fmt.Errorf(
					"web youtube takes 1 to %d URLs, got %d", nanogpt.MaxTranscribeURLs, len(args),
				)
			}

			client, err := cli.NanoClient(deps)
			if err != nil {
				return err
			}

			transcripts, err := client.TranscribeYouTube(cmd.Context(), args)
			if err != nil {
				return fmt.Errorf("web youtube: %w", err)
			}

			return printYouTube(deps, transcripts)
		},
	}
}

func printYouTube(deps cli.Deps, transcripts *nanogpt.YouTubeTranscripts) error {
	for index, transcript := range transcripts.Transcripts {
		if index > 0 {
			fmt.Fprintln(deps.Out)
		}

		fmt.Fprintf(deps.Out, "URL: %s\n", transcript.URL)
		fmt.Fprintf(deps.Out, "Success: %t\n", transcript.Success)

		if transcript.Title != "" {
			fmt.Fprintf(deps.Out, "Title: %s\n", transcript.Title)
		}

		if transcript.Transcript != "" {
			fmt.Fprintf(deps.Out, "Transcript:\n%s\n", transcript.Transcript)
		}

		if transcript.Error != "" {
			fmt.Fprintf(deps.Out, "Error: %s\n", transcript.Error)
		}
	}

	summary := transcripts.Summary
	fmt.Fprintf(deps.Out,
		"Summary: requested=%d processed=%d successful=%d failed=%d cost=%s\n",
		summary.Requested, summary.Processed, summary.Successful, summary.Failed,
		cli.FormatUSD(summary.TotalCost),
	)

	return nil
}
