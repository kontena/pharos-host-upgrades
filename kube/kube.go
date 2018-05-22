package kube

import (
	"fmt"

	"k8s.io/client-go/rest"
)

type Options struct {
	Namespace string
	DaemonSet string
	Node      string
}

type Kube struct {
	config  *rest.Config
	options Options
}

func New(options Options) (*Kube, error) {
	var kube = Kube{
		options: options,
	}

	if config, err := rest.InClusterConfig(); err != nil {
		return nil, fmt.Errorf("k8s.io/client-go/rest:InClusterConfig: %v", err)
	} else {
		kube.config = config
	}

	return &kube, nil
}
