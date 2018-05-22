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
	var kubeLock = kube.MakeLock(k)

	return kubeLock.With(f)
}
