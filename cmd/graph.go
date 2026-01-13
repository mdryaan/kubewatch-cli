package cmd

import (
	"context"
	"fmt"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	"github.com/mdryaan/kubewatch-cli/pkg/graph"
	"github.com/mdryaan/kubewatch-cli/pkg/output"
	"github.com/spf13/cobra"
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Show resource dependency graph",
	Long:  `Visualize which pods belong to which deployments and which services target which pods.`,
	RunE:  runGraph,
}

func runGraph(cmd *cobra.Command, args []string) error {
	kc, err := client.New(kubeconfig())
	if err != nil {
		return fmt.Errorf("connecting to cluster: %w", err)
	}

	ctx := context.Background()
	ns := namespace()
	if allNamespaces() {
		ns = ""
	}

	builder := graph.NewBuilder(kc)
	g, err := builder.Build(ctx, ns)
	if err != nil {
		return err
	}

	if len(g.Roots) == 0 {
		output.PrintInfo("no resources found in namespace: " + ns)
		return nil
	}

	output.PrintHeader(fmt.Sprintf("Resource Dependency Graph — %s", ns))

	if outputFormat() == output.FormatJSON {
		output.MustPrintJSON(g.Roots)
		return nil
	}

	graph.Render(g)
	fmt.Println()
	return nil
}
