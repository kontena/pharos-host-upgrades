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
	node *kube.Node
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

	if err := k.initNode(); err != nil {
		return k, err
	}

	return k, nil
}

func (k *Kube) initLock() error {
	if kubeLock, err := k.kube.Lock(); err != nil {
		return err
	} else if value, acquired, err := kubeLock.Test(); err != nil {
		return fmt.Errorf("Failed to test lock %v: %v", kubeLock, err)
	} else {
		log.Printf("Using kube lock %v (acquired=%v, value=%v)", kubeLock, acquired, value)

		k.lock = kubeLock
	}

	return nil
}

func (k *Kube) initNode() error {
	if kubeNode, err := k.kube.Node(); err != nil {
		return err
	} else if exists, err := kubeNode.HasCondition(UpgradeConditionType); err != nil {
		return fmt.Errorf("Failed to check node %v condition: %v", kubeNode, err)
	} else if exists {
		log.Printf("Using kube node %v", kubeNode)

		k.node = kubeNode
	} else if err := kubeNode.InitCondition(UpgradeConditionType); err != nil {
		return fmt.Errorf("Failed to initialize node %v condition: %v", kubeNode, err)
	} else {
		log.Printf("Iniitialized kube node %v", kubeNode)

		k.node = kubeNode
	}

	return nil
}

func (k Kube) WithLock(f func() error) error {
	if k.lock == nil {
		log.Printf("Skip kube locking")
		return f()
	}

	log.Printf("Acquiring kube lock...")

	return k.lock.With(f)
}

// Update node status condition based on function execution
func (k Kube) UpdateHostStatus(upgradeErr error) error {
	if k.node == nil {
		log.Printf("Skip kube node condition")
		return nil
	}

	log.Printf("Update kube node %v condition with error: %v", k.node, upgradeErr)

	if err := k.node.SetCondition(MakeUpgradeCondition(upgradeErr)); err != nil {
		log.Printf("Failed to update node %v condition: %v", k.node, err)
	}

	return nil
}
