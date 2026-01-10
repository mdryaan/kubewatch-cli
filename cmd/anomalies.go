package cmd

import (
	"context"
	"fmt"

	"github.com/mdryaan/kubewatch-cli/pkg/anomaly"
	"github.com/mdryaan/kubewatch-cli/pkg/client"
	"github.com/mdryaan/kubewatch-cli/pkg/output"
	"github.com/spf13/cobra"
)

var anomaliesCmd = &cobra.Command{
	Use:   "anomalies",
	Short: "Detect and list anomalous resources",
	Long:  `Scan for CrashLoopBackOff, OOMKilled, and Pending-too-long pods.`,
	RunE:  runAnomalies,
}

func init() {
	anomaliesCmd.Flags().StringSlice("severity", []string{"HIGH", "MEDIUM", "LOW"}, "filter by severity: HIGH, MEDIUM, LOW")
}

func runAnomalies(cmd *cobra.Command, args []string) error {
	kc, err := client.New(kubeconfig())
	if err != nil {
		return fmt.Errorf("connecting to cluster: %w", err)
	}

	ctx := context.Background()
	ns := namespace()
	if allNamespaces() {
		ns = ""
	}

	severities, _ := cmd.Flags().GetStringSlice("severity")
	severitySet := make(map[string]bool)
	for _, s := range severities {
		severitySet[s] = true
	}

	detector := anomaly.NewAnomalyDetector(kc)
	anomalies, err := detector.DetectAll(ctx, ns)
	if err != nil {
		return err
	}

	filtered := make([]anomaly.Anomaly, 0)
	for _, a := range anomalies {
		if severitySet[string(a.Severity)] {
			filtered = append(filtered, a)
		}
	}

	if len(filtered) == 0 {
		output.PrintSuccess("no anomalies detected in namespace: " + ns)
		return nil
	}

	output.PrintHeader(fmt.Sprintf("Anomalies Detected — %s  (%d found)", ns, len(filtered)))

	switch outputFormat() {
	case output.FormatJSON:
		output.MustPrintJSON(filtered)
	case output.FormatPlain:
		for _, a := range filtered {
			fmt.Printf("[%s] %s/%s: %s\n", a.Severity, a.ResourceKind, a.ResourceName, a.Message)
		}
	default:
		printAnomaliesTable(filtered)
	}
	return nil
}

func printAnomaliesTable(anomalies []anomaly.Anomaly) {
	t := output.NewTable([]string{"SEVERITY", "KIND", "NAME", "NAMESPACE", "TYPE", "MESSAGE", "AGE"})
	for _, a := range anomalies {
		severityColor := output.Warning(string(a.Severity))
		if a.Severity == anomaly.SeverityHigh {
			severityColor = output.Critical(string(a.Severity))
		}
		t.AddRow([]string{
			severityColor,
			output.Info(a.ResourceKind),
			a.ResourceName,
			a.Namespace,
			a.Type,
			a.Message,
			output.Faint(a.Age),
		})
	}
	t.Render()
	fmt.Println()
}
