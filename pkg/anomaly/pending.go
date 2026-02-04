package anomaly

import (
	"context"
	"fmt"
	"time"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	"github.com/mdryaan/kubewatch-cli/internal/config"
	"github.com/mdryaan/kubewatch-cli/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PendingDetector struct {
	client    *client.KubeClient
	threshold time.Duration
}

func NewPendingDetector(kc *client.KubeClient) *PendingDetector {
	return &PendingDetector{
		client:    kc,
		threshold: time.Duration(config.DefaultPendingThreshold) * time.Second,
	}
}

func (pd *PendingDetector) Detect(ctx context.Context, namespace string) ([]Anomaly, error) {
	pods, err := pd.client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing pods: %w", err)
	}

	var anomalies []Anomaly
	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodPending {
			continue
		}
		if !utils.IsOlderThan(pod.CreationTimestamp.Time, pd.threshold) {
			continue
		}

		reason := pendingReason(pod)
		anomalies = append(anomalies, Anomaly{
			ResourceKind: "Pod",
			ResourceName: pod.Name,
			Namespace:    pod.Namespace,
			Type:         "PendingTooLong",
			Severity:     SeverityMedium,
			Message:      fmt.Sprintf("pod pending for %s: %s", utils.Age(pod.CreationTimestamp.Time), reason),
			Age:          utils.Age(pod.CreationTimestamp.Time),
		})
	}
	return anomalies, nil
}

func pendingReason(pod corev1.Pod) string {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodScheduled && cond.Status == corev1.ConditionFalse {
			return cond.Message
		}
	}
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil {
			return cs.State.Waiting.Reason
		}
	}
	return "unknown reason"
}
