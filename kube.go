package main

import (
	"log"

	"github.com/kontena/pharos-host-upgrades/kube"
)

type KubeOptions struct {
	kube.Options
}

func (options KubeOptions) IsSet() bool {
	return !(options.Namespace == "" && options.DaemonSet == "" && options.Node == "")
}

type Kube struct {
	*kube.Kube
}

func makeKube(options Options) (Kube, error) {
	var k Kube

	if !options.Kube.IsSet() {
		log.Printf("Kube is not configured")
		return k, nil

	}

	if kube, err := kube.New(options.Kube.Options); err != nil {
		return k, err
	} else {
		k.Kube = kube
	}

	return k, nil
}

func (k Kube) withLock(f func() error) error {
	if k.Kube == nil {
		log.Printf("Skip kube locking")
		return f()
	}

	log.Printf("Acquiring kube lock...")

	if kubeLock, err := kube.NewLock(k.Kube); err != nil {
		return err
	} else {
		return kubeLock.With(f)
	}
}
