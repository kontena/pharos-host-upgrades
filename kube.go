package main

import (
	"fmt"
	"log"

	"github.com/kontena/pharos-host-upgrades/hosts"
	"github.com/kontena/pharos-host-upgrades/kube"
)

const KubeLockAnnotation = "pharos-host-upgrades.kontena.io/lock"

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
	host hosts.Host
}

func makeKube(options Options, host hosts.Host) (Kube, error) {
	var k = Kube{
		host: host,
	}

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
	if kubeLock, err := k.kube.Lock(KubeLockAnnotation); err != nil {
		return err
	} else {
		k.lock = kubeLock
	}

	if value, acquired, err := k.lock.Test(); err != nil {
		return fmt.Errorf("Failed to test lock %v: %v", k.lock, err)
	} else if !acquired {
		log.Printf("Using kube lock %v (not acquired, value=%v)", k.lock, value)
	} else if err := k.lock.Release(); err != nil {
		return fmt.Errorf("Failed to release lock %v: %v", k.lock, err)
	} else {
		log.Printf("Released kube lock %v (value=%v)", k.lock, value)
	}

	return nil
}

func (k *Kube) initNode() error {
	if kubeNode, err := k.kube.Node(); err != nil {
		return err
	} else {
		k.node = kubeNode
	}

	if exists, err := k.node.HasCondition(UpgradeConditionType); err != nil {
		return fmt.Errorf("Failed to check node %v condition: %v", k.node, err)
	} else if exists {
		log.Printf("Found kube node %v with existing conditions", k.node)
	} else if err := k.node.InitCondition(UpgradeConditionType); err != nil {
		return fmt.Errorf("Failed to initialize node %v condition %v: %v", k.node, UpgradeConditionType, err)
	} else if err := k.node.InitCondition(RebootConditionType); err != nil {
		return fmt.Errorf("Failed to initialize node %v condition %v: %v", k.node, RebootConditionType, err)
	} else {
		log.Printf("Initialized kube node %v conditions", k.node)
	}

	return nil
}

func (k Kube) AcquireLock() error {
	if k.lock == nil {
		log.Printf("Skip kube locking")
		return nil
	}

	log.Printf("Acquiring kube lock...")

	return k.lock.Acquire()
}

func (k Kube) ReleaseLock() error {
	if k.lock == nil {
		log.Printf("Skip kube locking")
		return nil
	}

	log.Printf("Releasing kube lock...")

	return k.lock.Release()
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
func (k Kube) UpdateHostStatus(status hosts.Status, upgradeErr error) error {
	if k.node == nil {
		log.Printf("Skip kube node condition")
		return nil
	}

	log.Printf("Update kube node %v condition for status=%v with error: %v", k.node, status, upgradeErr)

	if err := k.node.SetCondition(
		MakeUpgradeCondition(status, upgradeErr),
		MakeRebootCondition(status, k.host.Info(), upgradeErr),
	); err != nil {
		log.Printf("Failed to update node %v condition: %v", k.node, err)
	}

	return nil
}
