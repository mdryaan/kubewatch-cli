package watcher

import (
	"context"
	"time"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
)

type EventType string

const (
	EventAdded    EventType = "ADDED"
	EventModified EventType = "MODIFIED"
	EventDeleted  EventType = "DELETED"
)

type WatchEvent struct {
	Type      EventType
	Kind      string
	Name      string
	Namespace string
	Status    string
	Message   string
	Timestamp time.Time
}

type Watcher interface {
	Watch(ctx context.Context, namespace string, labelSelector string, events chan<- WatchEvent) error
}

type BaseWatcher struct {
	client *client.KubeClient
}

func newBaseWatcher(kc *client.KubeClient) BaseWatcher {
	return BaseWatcher{client: kc}
}
