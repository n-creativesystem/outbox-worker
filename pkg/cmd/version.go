package cmd

import (
	"fmt"

	"github.com/n-creativesystem/outbox-worker/pkg/internal/version"
	"github.com/spf13/cobra"
)

func versionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("version: %s\n", version.Version)
		},
	}
	return cmd
}
