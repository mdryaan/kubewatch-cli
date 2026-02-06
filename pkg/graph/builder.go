package graph

import (
	"context"
	"fmt"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeKind string

const (
	KindDeployment  NodeKind = "Deployment"
	KindReplicaSet  NodeKind = "ReplicaSet"
	KindPod         NodeKind = "Pod"
	KindService     NodeKind = "Service"
	KindStatefulSet NodeKind = "StatefulSet"
)

type GraphNode struct {
	Name      string
	Namespace string
	Kind      NodeKind
	Status    string
	Children  []*GraphNode
}

type Graph struct {
	Roots []*GraphNode
}

type Builder struct {
	client *client.KubeClient
}

func NewBuilder(kc *client.KubeClient) *Builder {
	return &Builder{client: kc}
}

func (b *Builder) Build(ctx context.Context, namespace string) (*Graph, error) {
	graph := &Graph{}

	deployments, err := b.client.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing deployments: %w", err)
	}

	replicaSets, err := b.client.Clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing replicasets: %w", err)
	}

	pods, err := b.client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing pods: %w", err)
	}

	services, err := b.client.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing services: %w", err)
	}

	rsMap := make(map[string]*GraphNode)
	for _, rs := range replicaSets.Items {
		node := &GraphNode{
			Name:      rs.Name,
			Namespace: rs.Namespace,
			Kind:      KindReplicaSet,
			Status:    fmt.Sprintf("%d/%d", rs.Status.ReadyReplicas, rs.Status.Replicas),
		}
		rsMap[rs.Name] = node
	}

	for _, pod := range pods.Items {
		podNode := &GraphNode{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Kind:      KindPod,
			Status:    string(pod.Status.Phase),
		}
		for _, owner := range pod.OwnerReferences {
			if owner.Kind == "ReplicaSet" {
				if rsNode, ok := rsMap[owner.Name]; ok {
					rsNode.Children = append(rsNode.Children, podNode)
				}
			}
		}
	}

	for _, d := range deployments.Items {
		depNode := &GraphNode{
			Name:      d.Name,
			Namespace: d.Namespace,
			Kind:      KindDeployment,
			Status:    fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, d.Status.Replicas),
		}
		for _, rs := range replicaSets.Items {
			for _, owner := range rs.OwnerReferences {
				if owner.Kind == "Deployment" && owner.Name == d.Name {
					if rsNode, ok := rsMap[rs.Name]; ok {
						depNode.Children = append(depNode.Children, rsNode)
					}
				}
			}
		}
		graph.Roots = append(graph.Roots, depNode)
	}

	for _, svc := range services.Items {
		if svc.Name == "kubernetes" {
			continue
		}
		svcNode := &GraphNode{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Kind:      KindService,
			Status:    string(svc.Spec.Type),
		}

		if svc.Spec.Selector != nil {
			for _, pod := range pods.Items {
				if labelsMatchSelector(pod.Labels, svc.Spec.Selector) {
					svcNode.Children = append(svcNode.Children, &GraphNode{
						Name:      pod.Name,
						Namespace: pod.Namespace,
						Kind:      KindPod,
						Status:    string(pod.Status.Phase),
					})
				}
			}
		}
		graph.Roots = append(graph.Roots, svcNode)
	}

	return graph, nil
}

func labelsMatchSelector(labels, selector map[string]string) bool {
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return len(selector) > 0
}
