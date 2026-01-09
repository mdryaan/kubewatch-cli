package cmd

import (
	"context"
	"fmt"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	"github.com/mdryaan/kubewatch-cli/pkg/output"
	"github.com/mdryaan/kubewatch-cli/pkg/summary"
	"github.com/spf13/cobra"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show namespace-level resource summary",
	Long:  `Display counts of pods, deployments, services, and overall health across namespaces.`,
	RunE:  runSummary,
}

func runSummary(cmd *cobra.Command, args []string) error {
	kc, err := client.New(kubeconfig())
	if err != nil {
		return fmt.Errorf("connecting to cluster: %w", err)
	}

	ctx := context.Background()
	collector := summary.NewCollector(kc)

	if allNamespaces() {
		summaries, err := collector.CollectAll(ctx)
		if err != nil {
			return err
		}
		if len(summaries) == 0 {
			output.PrintInfo("no namespaces found")
			return nil
		}
		summary.Report(summaries, outputFormat())
		return nil
	}

	ns := namespace()
	s, err := collector.Collect(ctx, ns)
	if err != nil {
		return err
	}

	summary.Report([]summary.NamespaceSummary{*s}, outputFormat())
	return nil
}
