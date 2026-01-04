package health

import (
	"context"
	"fmt"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	"github.com/mdryaan/kubewatch-cli/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceChecker struct {
	client *client.KubeClient
}

func NewServiceChecker(kc *client.KubeClient) *ServiceChecker {
	return &ServiceChecker{client: kc}
}

func (sc *ServiceChecker) Check(ctx context.Context, namespace string, labelSelector string) ([]ResourceHealth, error) {
	services, err := sc.client.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("listing services: %w", err)
	}

	results := make([]ResourceHealth, 0, len(services.Items))
	for _, svc := range services.Items {
		results = append(results, assessService(svc))
	}
	return results, nil
}

func assessService(svc corev1.Service) ResourceHealth {
	status, msg := serviceStatus(svc)
	return ResourceHealth{
		Name:      svc.Name,
		Namespace: svc.Namespace,
		Kind:      "Service",
		Status:    status,
		Message:   msg,
		Age:       utils.Age(svc.CreationTimestamp.Time),
		Labels:    svc.Labels,
	}
}

func serviceStatus(svc corev1.Service) (Status, string) {
	switch svc.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return StatusWarning, "waiting for external IP"
		}
		ingress := svc.Status.LoadBalancer.Ingress[0]
		addr := ingress.IP
		if addr == "" {
			addr = ingress.Hostname
		}
		return StatusHealthy, fmt.Sprintf("LoadBalancer: %s", addr)
	case corev1.ServiceTypeExternalName:
		return StatusHealthy, fmt.Sprintf("ExternalName: %s", svc.Spec.ExternalName)
	default:
		clusterIP := svc.Spec.ClusterIP
		if clusterIP == "" || clusterIP == "None" {
			return StatusHealthy, "headless service"
		}
		return StatusHealthy, fmt.Sprintf("ClusterIP: %s", clusterIP)
	}
}
