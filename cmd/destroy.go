package cmd

import (
	"bold/pkg/workflow"

	"github.com/spf13/cobra"
)

func NewDestroyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "destroy [manifest_file]",
		Short: "Menghancurkan (destroy) semua sumber daya dari sebuah layanan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return workflow.Run(args[0], "destroy")
		},
	}
}
