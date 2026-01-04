package health

import (
	"context"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
)

type Status string

const (
	StatusHealthy  Status = "Healthy"
	StatusWarning  Status = "Warning"
	StatusCritical Status = "Critical"
	StatusUnknown  Status = "Unknown"
)

type ResourceHealth struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Kind      string            `json:"kind"`
	Status    Status            `json:"status"`
	Message   string            `json:"message"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels,omitempty"`
}

type Checker interface {
	Check(ctx context.Context, namespace string, labelSelector string) ([]ResourceHealth, error)
}

type HealthChecker struct {
	client *client.KubeClient
	checkers []Checker
}

func NewHealthChecker(kc *client.KubeClient) *HealthChecker {
	return &HealthChecker{
		client: kc,
		checkers: []Checker{
			NewPodChecker(kc),
			NewDeploymentChecker(kc),
			NewNodeChecker(kc),
			NewServiceChecker(kc),
		},
	}
}

func (hc *HealthChecker) CheckAll(ctx context.Context, namespace string, labelSelector string) ([]ResourceHealth, error) {
	var all []ResourceHealth
	for _, c := range hc.checkers {
		results, err := c.Check(ctx, namespace, labelSelector)
		if err != nil {
			continue
		}
		all = append(all, results...)
	}
	return all, nil
}

func OverallStatus(items []ResourceHealth) Status {
	hasCritical := false
	hasWarning := false
	for _, item := range items {
		switch item.Status {
		case StatusCritical:
			hasCritical = true
		case StatusWarning:
			hasWarning = true
		}
	}
	if hasCritical {
		return StatusCritical
	}
	if hasWarning {
		return StatusWarning
	}
	return StatusHealthy
}
