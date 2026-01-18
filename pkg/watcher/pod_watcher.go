package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type PodWatcher struct {
	BaseWatcher
}

func NewPodWatcher(kc *client.KubeClient) *PodWatcher {
	return &PodWatcher{BaseWatcher: newBaseWatcher(kc)}
}

func (pw *PodWatcher) Watch(ctx context.Context, namespace string, labelSelector string, events chan<- WatchEvent) error {
	watcher, err := pw.client.Clientset.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return fmt.Errorf("starting pod watch: %w", err)
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
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}
			events <- WatchEvent{
				Type:      watchEventType(event.Type),
				Kind:      "Pod",
				Name:      pod.Name,
				Namespace: pod.Namespace,
				Status:    string(pod.Status.Phase),
				Message:   podReadiness(pod),
				Timestamp: time.Now(),
			}
		}
	}
}

func podReadiness(pod *corev1.Pod) string {
	if len(pod.Status.ContainerStatuses) == 0 {
		return string(pod.Status.Phase)
	}
	ready := 0
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			ready++
		}
	}
	return fmt.Sprintf("%d/%d", ready, len(pod.Status.ContainerStatuses))
}

func watchEventType(t watch.EventType) EventType {
	switch t {
	case watch.Added:
		return EventAdded
	case watch.Modified:
		return EventModified
	case watch.Deleted:
		return EventDeleted
	default:
		return EventModified
	}
}
