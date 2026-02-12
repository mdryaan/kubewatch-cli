package summary

import (
	"fmt"
	"os"

	"github.com/mdryaan/kubewatch-cli/pkg/output"
)

func Report(summaries []NamespaceSummary, format output.Format) {
	switch format {
	case output.FormatJSON:
		output.MustPrintJSON(summaries)
	case output.FormatPlain:
		reportPlain(summaries)
	default:
		reportTable(summaries)
	}
}

func reportTable(summaries []NamespaceSummary) {
	output.PrintHeader("Namespace Summary")

	t := output.NewTable([]string{"NAMESPACE", "PODS", "RUNNING", "PENDING", "FAILED", "DEPLOYMENTS", "SERVICES", "STATUS"})
	for _, s := range summaries {
		podStatus := healthStatus(s)
		t.AddRow([]string{
			output.Info(s.Namespace),
			fmt.Sprintf("%d", s.Pods.Total),
			output.Healthy(fmt.Sprintf("%d", s.Pods.Running)),
			output.Warning(fmt.Sprintf("%d", s.Pods.Pending)),
			output.Critical(fmt.Sprintf("%d", s.Pods.Failed)),
			fmt.Sprintf("%d/%d", s.Deployments.Available, s.Deployments.Total),
			fmt.Sprintf("%d", s.Services),
			output.StatusColor(podStatus),
		})
	}
	t.Render()
}

func reportPlain(summaries []NamespaceSummary) {
	for _, s := range summaries {
		fmt.Fprintf(os.Stdout, "Namespace: %s\n", s.Namespace)
		fmt.Fprintf(os.Stdout, "  Pods: %d total, %d running, %d pending, %d failed\n",
			s.Pods.Total, s.Pods.Running, s.Pods.Pending, s.Pods.Failed)
		fmt.Fprintf(os.Stdout, "  Deployments: %d/%d available\n",
			s.Deployments.Available, s.Deployments.Total)
		fmt.Fprintf(os.Stdout, "  Services: %d, ConfigMaps: %d, Secrets: %d\n\n",
			s.Services, s.ConfigMaps, s.Secrets)
	}
}

func healthStatus(s NamespaceSummary) string {
	if s.Pods.Failed > 0 {
		return "Critical"
	}
	if s.Pods.Pending > 0 || s.Deployments.Available < s.Deployments.Total {
		return "Warning"
	}
	return "Healthy"
}
