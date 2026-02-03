package anomaly

import (
	"context"
	"fmt"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	"github.com/mdryaan/kubewatch-cli/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OOMKillDetector struct {
	client *client.KubeClient
}

func NewOOMKillDetector(kc *client.KubeClient) *OOMKillDetector {
	return &OOMKillDetector{client: kc}
}

func (od *OOMKillDetector) Detect(ctx context.Context, namespace string) ([]Anomaly, error) {
	pods, err := od.client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing pods: %w", err)
	}

	var anomalies []Anomaly
	for _, pod := range pods.Items {
		for _, cs := range pod.Status.ContainerStatuses {
			isOOM := false
			restarts := cs.RestartCount

			if cs.State.Terminated != nil && cs.State.Terminated.Reason == "OOMKilled" {
				isOOM = true
			}
			if cs.LastTerminationState.Terminated != nil && cs.LastTerminationState.Terminated.Reason == "OOMKilled" {
				isOOM = true
			}

			if isOOM {
				anomalies = append(anomalies, Anomaly{
					ResourceKind: "Pod",
					ResourceName: pod.Name,
					Namespace:    pod.Namespace,
					Type:         "OOMKilled",
					Severity:     SeverityHigh,
					Message:      fmt.Sprintf("container '%s' was OOMKilled (restarts: %d) — increase memory limits", cs.Name, restarts),
					Age:          utils.Age(pod.CreationTimestamp.Time),
				})
				break
			}
		}
	}
	return anomalies, nil
}
