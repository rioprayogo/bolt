package cmd

import (
	"bold/pkg/workflow"

	"github.com/spf13/cobra"
)

func NewBootstrapCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "bootstrap [manifest_file]",
		Short: "Mem-bootstrap sebuah layanan (membuat atau memperbarui infrastruktur)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return workflow.Run(args[0], "apply")
		},
	}
}
