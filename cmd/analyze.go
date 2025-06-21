package cmd

import (
	"bold/pkg/cost"
	"bold/pkg/graph"
	"bold/pkg/parser"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func NewAnalyzeCommand() *cobra.Command {
	var format string
	var outputFile string

	cmd := &cobra.Command{
		Use:   "analyze [manifest_file]",
		Short: "Analyze infrastructure manifest (dependency graph & cost estimation)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestFile := args[0]

			service, err := parser.ParseManifest(manifestFile)
			if err != nil {
				return fmt.Errorf("failed to parse manifest: %w", err)
			}

			fmt.Println("🔍 Analyzing infrastructure manifest...")
			fmt.Printf("Service: %s (Owner: %s)\n", service.Metadata.Name, service.Metadata.Owner)
			fmt.Printf("Providers: %d\n\n", len(service.Providers))

			dependencyGraph := graph.GenerateDependencyGraph(service)
			costReport := cost.EstimateCosts(service)

			var output string

			switch format {
			case "tree":
				output = graph.PrintDependencyTree(dependencyGraph)
			case "mermaid":
				output = graph.GenerateMermaidDiagram(dependencyGraph)
			case "dot":
				output = graph.GenerateDotGraph(dependencyGraph)
			case "cost":
				output = cost.FormatCostReport(costReport)
			case "full":
				output = generateFullAnalysis(dependencyGraph, costReport)
			default:
				output = generateFullAnalysis(dependencyGraph, costReport)
			}

			if outputFile != "" {
				err := os.WriteFile(outputFile, []byte(output), 0644)
				if err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
				fmt.Printf("✅ Analysis saved to: %s\n", outputFile)
			} else {
				fmt.Println(output)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "full", "Output format (tree, mermaid, dot, cost, full)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (optional)")

	return cmd
}

func generateFullAnalysis(dependencyGraph *graph.DependencyGraph, costReport *cost.CostReport) string {
	var analysis strings.Builder

	analysis.WriteString("🚀 BOLT INFRASTRUCTURE ANALYSIS\n")
	analysis.WriteString("================================\n\n")

	analysis.WriteString("📊 DEPENDENCY GRAPH\n")
	analysis.WriteString("-------------------\n")
	analysis.WriteString(graph.PrintDependencyTree(dependencyGraph))
	analysis.WriteString("\n")

	analysis.WriteString("💰 COST ESTIMATION\n")
	analysis.WriteString("------------------\n")
	analysis.WriteString(cost.FormatCostReport(costReport))

	return analysis.String()
}
