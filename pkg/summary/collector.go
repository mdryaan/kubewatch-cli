package summary

import (
	"context"
	"fmt"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NamespaceSummary struct {
	Namespace   string          `json:"namespace"`
	Pods        PodSummary      `json:"pods"`
	Deployments DeploymentSummary `json:"deployments"`
	Services    int             `json:"services"`
	ConfigMaps  int             `json:"configmaps"`
	Secrets     int             `json:"secrets"`
}

type PodSummary struct {
	Total   int `json:"total"`
	Running int `json:"running"`
	Pending int `json:"pending"`
	Failed  int `json:"failed"`
}

type DeploymentSummary struct {
	Total     int `json:"total"`
	Available int `json:"available"`
}

type Collector struct {
	client *client.KubeClient
}

func NewCollector(kc *client.KubeClient) *Collector {
	return &Collector{client: kc}
}

func (c *Collector) Collect(ctx context.Context, namespace string) (*NamespaceSummary, error) {
	summary := &NamespaceSummary{Namespace: namespace}

	pods, err := c.client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing pods: %w", err)
	}
	for _, pod := range pods.Items {
		summary.Pods.Total++
		switch pod.Status.Phase {
		case corev1.PodRunning:
			summary.Pods.Running++
		case corev1.PodPending:
			summary.Pods.Pending++
		case corev1.PodFailed:
			summary.Pods.Failed++
		}
	}

	deployments, err := c.client.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing deployments: %w", err)
	}
	for _, d := range deployments.Items {
		summary.Deployments.Total++
		desired := int32(1)
		if d.Spec.Replicas != nil {
			desired = *d.Spec.Replicas
		}
		if d.Status.AvailableReplicas >= desired {
			summary.Deployments.Available++
		}
	}

	services, err := c.client.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing services: %w", err)
	}
	summary.Services = len(services.Items)

	cms, err := c.client.Clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing configmaps: %w", err)
	}
	summary.ConfigMaps = len(cms.Items)

	secrets, err := c.client.Clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing secrets: %w", err)
	}
	summary.Secrets = len(secrets.Items)

	return summary, nil
}

func (c *Collector) CollectAll(ctx context.Context) ([]NamespaceSummary, error) {
	namespaces, err := c.client.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing namespaces: %w", err)
	}

	var summaries []NamespaceSummary
	for _, ns := range namespaces.Items {
		s, err := c.Collect(ctx, ns.Name)
		if err != nil {
			continue
		}
		summaries = append(summaries, *s)
	}
	return summaries, nil
}
