package health

import (
	"context"
	"fmt"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	"github.com/mdryaan/kubewatch-cli/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeChecker struct {
	client *client.KubeClient
}

func NewNodeChecker(kc *client.KubeClient) *NodeChecker {
	return &NodeChecker{client: kc}
}

func (nc *NodeChecker) Check(ctx context.Context, namespace string, labelSelector string) ([]ResourceHealth, error) {
	nodes, err := nc.client.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("listing nodes: %w", err)
	}

	results := make([]ResourceHealth, 0, len(nodes.Items))
	for _, n := range nodes.Items {
		results = append(results, assessNode(n))
	}
	return results, nil
}

func assessNode(node corev1.Node) ResourceHealth {
	status, msg := nodeStatus(node)
	return ResourceHealth{
		Name:      node.Name,
		Namespace: "",
		Kind:      "Node",
		Status:    status,
		Message:   msg,
		Age:       utils.Age(node.CreationTimestamp.Time),
		Labels:    node.Labels,
	}
}

func nodeStatus(node corev1.Node) (Status, string) {
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			if cond.Status == corev1.ConditionTrue {
				if node.Spec.Unschedulable {
					return StatusWarning, "node is cordoned"
				}
				return StatusHealthy, "node ready"
			}
			return StatusCritical, fmt.Sprintf("node not ready: %s", cond.Message)
		}
	}

	for _, cond := range node.Status.Conditions {
		switch cond.Type {
		case corev1.NodeMemoryPressure:
			if cond.Status == corev1.ConditionTrue {
				return StatusCritical, "node has memory pressure"
			}
		case corev1.NodeDiskPressure:
			if cond.Status == corev1.ConditionTrue {
				return StatusCritical, "node has disk pressure"
			}
		case corev1.NodePIDPressure:
			if cond.Status == corev1.ConditionTrue {
				return StatusCritical, "node has PID pressure"
			}
		}
	}

	return StatusUnknown, "unknown node status"
}
