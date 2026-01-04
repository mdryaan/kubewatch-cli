package health

import (
	"context"
	"fmt"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	"github.com/mdryaan/kubewatch-cli/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodChecker struct {
	client *client.KubeClient
}

func NewPodChecker(kc *client.KubeClient) *PodChecker {
	return &PodChecker{client: kc}
}

func (pc *PodChecker) Check(ctx context.Context, namespace string, labelSelector string) ([]ResourceHealth, error) {
	pods, err := pc.client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("listing pods: %w", err)
	}

	results := make([]ResourceHealth, 0, len(pods.Items))
	for _, pod := range pods.Items {
		results = append(results, assessPod(pod))
	}
	return results, nil
}

func assessPod(pod corev1.Pod) ResourceHealth {
	status, msg := podStatus(pod)
	return ResourceHealth{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Kind:      "Pod",
		Status:    status,
		Message:   msg,
		Age:       utils.Age(pod.CreationTimestamp.Time),
		Labels:    pod.Labels,
	}
}

func podStatus(pod corev1.Pod) (Status, string) {
	if pod.DeletionTimestamp != nil {
		return StatusWarning, "Terminating"
	}

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil {
			reason := cs.State.Waiting.Reason
			switch reason {
			case "CrashLoopBackOff":
				return StatusCritical, fmt.Sprintf("container %s is in CrashLoopBackOff", cs.Name)
			case "OOMKilled":
				return StatusCritical, fmt.Sprintf("container %s was OOMKilled", cs.Name)
			case "ImagePullBackOff", "ErrImagePull":
				return StatusCritical, fmt.Sprintf("container %s cannot pull image", cs.Name)
			case "ContainerCreating":
				return StatusWarning, "container creating"
			}
		}
		if cs.State.Terminated != nil && cs.State.Terminated.Reason == "OOMKilled" {
			return StatusCritical, fmt.Sprintf("container %s was OOMKilled", cs.Name)
		}
	}

	switch pod.Status.Phase {
	case corev1.PodRunning:
		readyCount := 0
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Ready {
				readyCount++
			}
		}
		total := len(pod.Status.ContainerStatuses)
		if readyCount < total {
			return StatusWarning, fmt.Sprintf("%d/%d containers ready", readyCount, total)
		}
		return StatusHealthy, fmt.Sprintf("%d/%d containers ready", readyCount, total)
	case corev1.PodPending:
		return StatusWarning, "pod is pending"
	case corev1.PodFailed:
		return StatusCritical, "pod failed"
	case corev1.PodSucceeded:
		return StatusHealthy, "completed"
	}

	return StatusUnknown, string(pod.Status.Phase)
}
