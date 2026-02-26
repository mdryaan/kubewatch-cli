package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EventWatcher struct {
	BaseWatcher
}

func NewEventWatcher(kc *client.KubeClient) *EventWatcher {
	return &EventWatcher{BaseWatcher: newBaseWatcher(kc)}
}

func (ew *EventWatcher) Watch(ctx context.Context, namespace string, labelSelector string, events chan<- WatchEvent) error {
	w, err := ew.client.Clientset.CoreV1().Events(namespace).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("starting event watch: %w", err)
	}
	defer w.Stop()

	ch := w.ResultChan()
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-ch:
			if !ok {
				return nil
			}
			ev, ok := event.Object.(*corev1.Event)
			if !ok {
				continue
			}
			events <- WatchEvent{
				Type:      EventType(ev.Type),
				Kind:      ev.InvolvedObject.Kind,
				Name:      ev.InvolvedObject.Name,
				Namespace: ev.Namespace,
				Status:    ev.Reason,
				Message:   ev.Message,
				Timestamp: eventTimestamp(ev),
			}
		}
	}
}

func eventTimestamp(ev *corev1.Event) time.Time {
	if !ev.LastTimestamp.IsZero() {
		return ev.LastTimestamp.Time
	}
	if !ev.FirstTimestamp.IsZero() {
		return ev.FirstTimestamp.Time
	}
	return time.Now()
}
