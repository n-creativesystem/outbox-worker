package cmd

import "github.com/spf13/cobra"

func rootCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:           "outbox-messenger",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(outboxCommand())
	cmd.AddCommand(versionCommand())

	return &cmd
}
