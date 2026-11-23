package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wishmatic/nano-mcp/internal/cli"
	"github.com/wishmatic/nano-mcp/internal/nanogpt"
)

func newGenVideoCommand(deps cli.Deps) *cobra.Command {
	var (
		model          string
		image          string
		negativePrompt string
		duration       string
		aspectRatio    string
		resolution     string
		mode           string
		videoURL       string
		out            string
	)

	cmd := &cobra.Command{
		Use:   "video <prompt...>",
		Short: "Generate a video through NanoGPT",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt, err := promptFromArgs(args)
			if err != nil {
				return err
			}

			req := nanogpt.VideoRequest{
				Model:          model,
				Prompt:         prompt,
				NegativePrompt: negativePrompt,
				Duration:       duration,
				AspectRatio:    aspectRatio,
				Resolution:     resolution,
				Mode:           mode,
				Seed:           cli.OptionalInt(cmd, "seed"),
				GenerateAudio:  cli.OptionalBool(cmd, "generate-audio"),
				VideoURL:       videoURL,
			}

			if image != "" {
				resolver, err := cli.ResolverFromEnv()
				if err != nil {
					return err
				}

				imageDataURL, err := cli.ResolveDataURL(cmd.Context(), resolver, image)
				if err != nil {
					return fmt.Errorf("gen video: %w", err)
				}

				req.ImageDataURL = imageDataURL

				if req.Mode == "" {
					req.Mode = nanogpt.ModeImageToVideo
				}
			}

			client, err := cli.NanoClient(deps)
			if err != nil {
				return err
			}

			video, err := client.GenerateVideo(cmd.Context(), req)
			if err != nil {
				return fmt.Errorf("gen video: %w", err)
			}

			data, mediaType, err := cli.DownloadAsset(cmd.Context(), video.URL, "video/mp4")
			if err != nil {
				return fmt.Errorf("gen video: %w", err)
			}

			path, err := cli.WriteAsset(out, "nano-video", data, mediaType)
			if err != nil {
				return fmt.Errorf("gen video: %w", err)
			}

			fmt.Fprintln(deps.Out, path)
			fmt.Fprintf(deps.Out, "Cost: %s\n", cli.FormatUSD(video.CostUSD))

			return nil
		},
	}

	cmd.Flags().StringVar(&model, "model", "", "video model id")
	cmd.Flags().StringVar(&image, "image", "", "image to animate")
	cmd.Flags().StringVar(&negativePrompt, "negative-prompt", "", "content to suppress")
	cmd.Flags().StringVar(&duration, "duration", "", "length in seconds")
	cmd.Flags().StringVar(&aspectRatio, "aspect-ratio", "", "16:9, 9:16, 1:1, 4:3, or 3:4")
	cmd.Flags().StringVar(&resolution, "resolution", "", "480p, 720p, 1080p, 2k, or 4k")
	cmd.Flags().StringVar(
		&mode, "mode", "", "text-to-video, image-to-video, reference-to-video, or video-edit",
	)
	cmd.Flags().Int("seed", 0, "seed for the providers that honour one")
	cmd.Flags().Bool("generate-audio", false, "add AI audio on the models that support it")
	cmd.Flags().StringVar(&videoURL, "video-url", "", "source video to extend, edit, or upscale")
	cmd.Flags().StringVar(&out, "out", ".", "directory for the generated video")

	_ = cmd.MarkFlagRequired("model")

	return cmd
}
