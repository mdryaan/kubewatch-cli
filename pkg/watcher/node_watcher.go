package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeWatcher struct {
	BaseWatcher
}

func NewNodeWatcher(kc *client.KubeClient) *NodeWatcher {
	return &NodeWatcher{BaseWatcher: newBaseWatcher(kc)}
}

func (nw *NodeWatcher) Watch(ctx context.Context, namespace string, labelSelector string, events chan<- WatchEvent) error {
	watcher, err := nw.client.Clientset.CoreV1().Nodes().Watch(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return fmt.Errorf("starting node watch: %w", err)
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return nil
			}
			node, ok := event.Object.(*corev1.Node)
			if !ok {
				continue
			}
			events <- WatchEvent{
				Type:      watchEventType(event.Type),
				Kind:      "Node",
				Name:      node.Name,
				Namespace: "",
				Status:    nodeReadyStatus(node),
				Message:   nodeRoles(node),
				Timestamp: time.Now(),
			}
		}
	}
}

func nodeReadyStatus(node *corev1.Node) string {
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			if cond.Status == corev1.ConditionTrue {
				return "Ready"
			}
			return "NotReady"
		}
	}
	return "Unknown"
}

func nodeRoles(node *corev1.Node) string {
	roles := ""
	for k := range node.Labels {
		if k == "node-role.kubernetes.io/control-plane" || k == "node-role.kubernetes.io/master" {
			roles = "control-plane"
		}
		if k == "node-role.kubernetes.io/worker" {
			roles = "worker"
		}
	}
	if roles == "" {
		return "worker"
	}
	return roles
}
