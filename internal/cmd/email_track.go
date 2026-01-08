package cmd

import (
	"github.com/spf13/cobra"
)

func newEmailTrackCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "track",
		Short: "Email open tracking",
		Long:  `Commands for setting up and querying email open tracking.`,
	}

	cmd.AddCommand(newEmailTrackSetupCmd(flags))
	cmd.AddCommand(newEmailTrackOpensCmd(flags))
	cmd.AddCommand(newEmailTrackStatusCmd(flags))

	return cmd
}
