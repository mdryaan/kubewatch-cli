package anomaly

import (
	"context"
	"fmt"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	"github.com/mdryaan/kubewatch-cli/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CrashLoopDetector struct {
	client *client.KubeClient
}

func NewCrashLoopDetector(kc *client.KubeClient) *CrashLoopDetector {
	return &CrashLoopDetector{client: kc}
}

func (cd *CrashLoopDetector) Detect(ctx context.Context, namespace string) ([]Anomaly, error) {
	pods, err := cd.client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing pods: %w", err)
	}

	var anomalies []Anomaly
	for _, pod := range pods.Items {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CrashLoopBackOff" {
				anomalies = append(anomalies, Anomaly{
					ResourceKind: "Pod",
					ResourceName: pod.Name,
					Namespace:    pod.Namespace,
					Type:         "CrashLoopBackOff",
					Severity:     SeverityHigh,
					Message:      fmt.Sprintf("container '%s' is in CrashLoopBackOff (restarts: %d)", cs.Name, cs.RestartCount),
					Age:          utils.Age(pod.CreationTimestamp.Time),
				})
				break
			}
		}
	}
	return anomalies, nil
}
