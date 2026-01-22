package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceWatcher struct {
	BaseWatcher
}

func NewServiceWatcher(kc *client.KubeClient) *ServiceWatcher {
	return &ServiceWatcher{BaseWatcher: newBaseWatcher(kc)}
}

func (sw *ServiceWatcher) Watch(ctx context.Context, namespace string, labelSelector string, events chan<- WatchEvent) error {
	watcher, err := sw.client.Clientset.CoreV1().Services(namespace).Watch(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return fmt.Errorf("starting service watch: %w", err)
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
			svc, ok := event.Object.(*corev1.Service)
			if !ok {
				continue
			}
			events <- WatchEvent{
				Type:      watchEventType(event.Type),
				Kind:      "Service",
				Name:      svc.Name,
				Namespace: svc.Namespace,
				Status:    string(svc.Spec.Type),
				Message:   serviceIP(svc),
				Timestamp: time.Now(),
			}
		}
	}
}

func serviceIP(svc *corev1.Service) string {
	if svc.Spec.ClusterIP != "" {
		return svc.Spec.ClusterIP
	}
	return "None"
}
