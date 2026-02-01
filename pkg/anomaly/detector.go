package anomaly

import (
	"context"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
)

type Severity string

const (
	SeverityHigh   Severity = "HIGH"
	SeverityMedium Severity = "MEDIUM"
	SeverityLow    Severity = "LOW"
)

type Anomaly struct {
	ResourceKind string   `json:"resource_kind"`
	ResourceName string   `json:"resource_name"`
	Namespace    string   `json:"namespace"`
	Type         string   `json:"type"`
	Severity     Severity `json:"severity"`
	Message      string   `json:"message"`
	Age          string   `json:"age"`
}

type AnomalyDetector struct {
	client     *client.KubeClient
	detectors  []detector
}

type detector interface {
	Detect(ctx context.Context, namespace string) ([]Anomaly, error)
}

func NewAnomalyDetector(kc *client.KubeClient) *AnomalyDetector {
	return &AnomalyDetector{
		client: kc,
		detectors: []detector{
			NewCrashLoopDetector(kc),
			NewOOMKillDetector(kc),
			NewPendingDetector(kc),
		},
	}
}

func (ad *AnomalyDetector) DetectAll(ctx context.Context, namespace string) ([]Anomaly, error) {
	var all []Anomaly
	for _, d := range ad.detectors {
		results, err := d.Detect(ctx, namespace)
		if err != nil {
			continue
		}
		all = append(all, results...)
	}
	return all, nil
}
