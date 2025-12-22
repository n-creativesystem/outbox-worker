package cmd

import (
	"github.com/spf13/cobra"
)

func outboxCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:           "outbox",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(outboxPollingCommand())
	return &cmd
}
