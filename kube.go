package main

import (
	"fmt"
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
	kube *kube.Kube
	lock *kube.Lock
}

func makeKube(options Options) (Kube, error) {
	var k Kube

	if !options.Kube.IsSet() {
		log.Printf("No --kube configuration")
		return k, nil
	}

	log.Printf("Using --kube-namespace=%v --kube-daemonset=%v --kube-node=%v",
		options.Kube.Namespace,
		options.Kube.DaemonSet,
		options.Kube.Node,
	)

	if kube, err := kube.New(options.Kube.Options); err != nil {
		return k, err
	} else {
		k.kube = kube
	}

	if err := k.initLock(); err != nil {
		return k, err
	}

	return k, nil
}

func (k *Kube) initLock() error {
	if kubeLock, err := kube.NewLock(k.kube); err != nil {
		return err
	} else if _, err := kubeLock.Test(); err != nil {
		return fmt.Errorf("Failed to test lock %v: %v", kubeLock, err)
	} else {
		k.lock = kubeLock
	}

	return nil
}

func (k Kube) withLock(f func() error) error {
	if k.lock == nil {
		log.Printf("Skip kube locking")
		return f()
	}

	log.Printf("Acquiring kube lock...")

	return k.lock.With(f)
}
