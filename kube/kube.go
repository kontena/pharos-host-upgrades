package kube

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Options struct {
	Namespace string
	DaemonSet string
	Node      string
}

type Kube struct {
	config  *rest.Config
	client  *kubernetes.Clientset
	options Options
}

func New(options Options) (*Kube, error) {
	var kube = Kube{
		options: options,
	}

	if kubeConfig, err := rest.InClusterConfig(); err != nil {
		return nil, fmt.Errorf("k8s.io/client-go/rest:InClusterConfig: %v", err)
	} else if client, err := kubernetes.NewForConfig(kubeConfig); err != nil {
		return nil, fmt.Errorf("k8s.io/client-go/kubernetes:NewForConfig: %v", err)
	} else {
		kube.client = client
	}

	return &kube, nil
}
