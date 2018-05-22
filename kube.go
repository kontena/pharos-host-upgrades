package main

import (
	"github.com/kontena/pharos-host-upgrades/kube"
)

type KubeOptions struct {
	kube.Options
}

func newKube(options Options) (*kube.Kube, error) {
	return kube.New(options.Kube.Options)
}

func withKubeLock(k *kube.Kube, f func() error) error {
	if kubeLock, err := kube.MakeLock(k); err != nil {
		return err
	} else {
		return kubeLock.With(f)
	}
}
