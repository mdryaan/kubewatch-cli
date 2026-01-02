package client

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubeClient struct {
	Clientset  kubernetes.Interface
	RestConfig *rest.Config
	Kubeconfig string
}

func New(kubeconfigPath string) (*KubeClient, error) {
	resolved := resolveKubeconfig(kubeconfigPath)
	cfg, err := buildRestConfig(resolved)
	if err != nil {
		return nil, fmt.Errorf("failed to build rest config: %w", err)
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &KubeClient{
		Clientset:  cs,
		RestConfig: cfg,
		Kubeconfig: resolved,
	}, nil
}

func (k *KubeClient) ServerVersion() (string, error) {
	info, err := k.Clientset.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return info.GitVersion, nil
}
