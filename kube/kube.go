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

func (kube *Kube) Node() (*Node, error) {
	var node = Node{
		name: kube.options.Node,
	}

	if err := node.connect(kube.config); err != nil {
		return nil, err
	}

	return &node, nil
}

func (kube *Kube) Lock() (*Lock, error) {
	var lock = Lock{
		namespace:  kube.options.Namespace,
		name:       kube.options.DaemonSet,
		annotation: LockAnnotation,
		value:      kube.options.Node,
	}

	if err := lock.connect(kube.config); err != nil {
		return nil, err
	}

	return &lock, nil
}
