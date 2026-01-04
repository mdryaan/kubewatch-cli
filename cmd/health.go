package cmd

import (
	"context"
	"fmt"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	"github.com/mdryaan/kubewatch-cli/pkg/health"
	"github.com/mdryaan/kubewatch-cli/pkg/output"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Show health status of all resources in a namespace",
	Long:  `Check the health of pods, deployments, services, and nodes in the target namespace.`,
	RunE:  runHealth,
}

func init() {
	healthCmd.Flags().StringSliceP("kinds", "k", []string{"pod", "deployment", "service", "node"}, "resource types to check")
}

func runHealth(cmd *cobra.Command, args []string) error {
	kc, err := client.New(kubeconfig())
	if err != nil {
		return fmt.Errorf("connecting to cluster: %w", err)
	}

	ctx := context.Background()
	ns := namespace()
	if allNamespaces() {
		ns = ""
	}

	checker := health.NewHealthChecker(kc)
	results, err := checker.CheckAll(ctx, ns, labelSelector())
	if err != nil {
		return err
	}

	if len(results) == 0 {
		output.PrintInfo("no resources found in namespace: " + ns)
		return nil
	}

	overall := health.OverallStatus(results)
	output.PrintHeader(fmt.Sprintf("Health Report — %s  [%s]", ns, string(overall)))

	switch outputFormat() {
	case output.FormatJSON:
		output.MustPrintJSON(results)
	case output.FormatPlain:
		for _, r := range results {
			fmt.Printf("%-12s %-40s %-10s %s\n", r.Kind, r.Name, r.Status, r.Message)
		}
	default:
		printHealthTable(results)
	}
	return nil
}

func printHealthTable(results []health.ResourceHealth) {
	t := output.NewTable([]string{"KIND", "NAME", "NAMESPACE", "STATUS", "MESSAGE", "AGE"})
	for _, r := range results {
		t.AddRow([]string{
			output.Info(r.Kind),
			r.Name,
			r.Namespace,
			output.StatusColor(string(r.Status)),
			r.Message,
			output.Faint(r.Age),
		})
	}
	t.Render()
	fmt.Println()
}
