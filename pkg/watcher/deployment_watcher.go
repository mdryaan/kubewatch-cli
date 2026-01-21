package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentWatcher struct {
	BaseWatcher
}

func NewDeploymentWatcher(kc *client.KubeClient) *DeploymentWatcher {
	return &DeploymentWatcher{BaseWatcher: newBaseWatcher(kc)}
}

func (dw *DeploymentWatcher) Watch(ctx context.Context, namespace string, labelSelector string, events chan<- WatchEvent) error {
	watcher, err := dw.client.Clientset.AppsV1().Deployments(namespace).Watch(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return fmt.Errorf("starting deployment watch: %w", err)
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
			d, ok := event.Object.(*appsv1.Deployment)
			if !ok {
				continue
			}
			events <- WatchEvent{
				Type:      watchEventType(event.Type),
				Kind:      "Deployment",
				Name:      d.Name,
				Namespace: d.Namespace,
				Status:    deploymentReadiness(d),
				Message:   deploymentMessage(d),
				Timestamp: time.Now(),
			}
		}
	}
}

func deploymentReadiness(d *appsv1.Deployment) string {
	desired := int32(1)
	if d.Spec.Replicas != nil {
		desired = *d.Spec.Replicas
	}
	return fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, desired)
}

func deploymentMessage(d *appsv1.Deployment) string {
	for _, cond := range d.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable {
			return string(cond.Status)
		}
	}
	return "Unknown"
}
