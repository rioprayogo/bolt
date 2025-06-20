package main

import (
	"bold/cmd"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "bolt",
		Short: "Bolt adalah tool untuk merakit infrastruktur dari definisi layanan.",
		Long:  `Bolt mengambil definisi layanan abstrak dan mensintesisnya menjadi infrastruktur nyata menggunakan engine seperti OpenTofu.`,
	}
	rootCmd.AddCommand(cmd.NewBootstrapCommand())
	rootCmd.AddCommand(cmd.NewDestroyCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
