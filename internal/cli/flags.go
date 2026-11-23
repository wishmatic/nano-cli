package cli

import "github.com/spf13/cobra"

func OptionalBool(cmd *cobra.Command, name string) *bool {
	if !cmd.Flags().Changed(name) {
		return nil
	}

	value, _ := cmd.Flags().GetBool(name)

	return &value
}
