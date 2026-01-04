package health

import (
	"context"
	"fmt"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	"github.com/mdryaan/kubewatch-cli/internal/utils"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentChecker struct {
	client *client.KubeClient
}

func NewDeploymentChecker(kc *client.KubeClient) *DeploymentChecker {
	return &DeploymentChecker{client: kc}
}

func (dc *DeploymentChecker) Check(ctx context.Context, namespace string, labelSelector string) ([]ResourceHealth, error) {
	deployments, err := dc.client.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("listing deployments: %w", err)
	}

	results := make([]ResourceHealth, 0, len(deployments.Items))
	for _, d := range deployments.Items {
		results = append(results, assessDeployment(d))
	}
	return results, nil
}

func assessDeployment(d appsv1.Deployment) ResourceHealth {
	status, msg := deploymentStatus(d)
	return ResourceHealth{
		Name:      d.Name,
		Namespace: d.Namespace,
		Kind:      "Deployment",
		Status:    status,
		Message:   msg,
		Age:       utils.Age(d.CreationTimestamp.Time),
		Labels:    d.Labels,
	}
}

func deploymentStatus(d appsv1.Deployment) (Status, string) {
	desired := int32(1)
	if d.Spec.Replicas != nil {
		desired = *d.Spec.Replicas
	}

	ready := d.Status.ReadyReplicas
	updated := d.Status.UpdatedReplicas
	available := d.Status.AvailableReplicas

	if desired == 0 {
		return StatusWarning, "scaled to zero"
	}

	if ready == desired && updated == desired && available == desired {
		return StatusHealthy, fmt.Sprintf("%d/%d replicas ready", ready, desired)
	}

	for _, cond := range d.Status.Conditions {
		if cond.Type == appsv1.DeploymentProgressing && cond.Status == "False" {
			return StatusCritical, cond.Message
		}
		if cond.Type == appsv1.DeploymentAvailable && cond.Status == "False" {
			return StatusCritical, "deployment unavailable"
		}
	}

	if ready < desired {
		return StatusWarning, fmt.Sprintf("%d/%d replicas ready", ready, desired)
	}

	return StatusWarning, fmt.Sprintf("rolling update in progress: %d/%d", updated, desired)
}
