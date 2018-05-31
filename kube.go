package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kontena/pharos-host-upgrades/hosts"
	"github.com/kontena/pharos-host-upgrades/kube"
	"github.com/kontena/pharos-host-upgrades/kubectl"
)

const KubeLockAnnotation = "pharos-host-upgrades.kontena.io/lock"
const KubeDrainAnnotation = "pharos-host-upgrades.kontena.io/drain"
const KubeRebootAnnotation = "pharos-host-upgrades.kontena.io/reboot"

type KubeOptions struct {
	kube.Options
}

func (options KubeOptions) IsSet() bool {
	return !(options.Namespace == "" && options.DaemonSet == "" && options.Node == "")
}

type Kube struct {
	options  kube.Options
	hostInfo hosts.Info
	kube     *kube.Kube
	lock     *kube.Lock
	node     *kube.Node
}

func makeKube(options Options, hostInfo hosts.Info) (*Kube, error) {
	var k = Kube{
		options:  options.Kube.Options,
		hostInfo: hostInfo,
	}

	if !options.Kube.IsSet() {
		log.Printf("No --kube configuration")
		return nil, nil
	}

	log.Printf("Using --kube-namespace=%v --kube-daemonset=%v --kube-node=%v",
		options.Kube.Namespace,
		options.Kube.DaemonSet,
		options.Kube.Node,
	)

	if kube, err := kube.New(options.Kube.Options); err != nil {
		return nil, err
	} else {
		k.kube = kube
	}

	if err := k.initNode(); err != nil {
		return nil, err
	}

	if err := k.initLock(); err != nil {
		return nil, err
	}

	// verifies host <=> node state, fails if not rebooted
	if err := k.clearNodeReboot(); err != nil {
		return nil, err
	}

	// this happens even without the reboot annotation set, we do not want to leave the node drained in case of errors
	if err := k.clearNodeDrain(); err != nil {
		return nil, err
	}

	// clear lock if acquired, assuming that host is now in a good state (rebooted, undrained)
	if err := k.clearLock(); err != nil {
		return nil, err
	}

	return &k, nil
}

func (k *Kube) initNode() error {
	if kubeNode, err := k.kube.Node(); err != nil {
		return err
	} else {
		k.node = kubeNode
	}

	return nil
}

func (k *Kube) initLock() error {
	if kubeLock, err := k.kube.Lock(KubeLockAnnotation); err != nil {
		return err
	} else {
		k.lock = kubeLock
	}

	return nil
}

func (k *Kube) checkReboot() (time.Time, bool, error) {
	var t time.Time

	if value, exists, err := k.node.GetAnnotation(KubeRebootAnnotation); err != nil {
		return t, false, fmt.Errorf("Faield to get node reboot annotation: %v", err)
	} else if !exists {
		return t, false, nil
	} else if err := json.Unmarshal([]byte(value), &t); err != nil {
		return t, true, fmt.Errorf("Failed to unmarshal reboot annotation: %v", err)
	} else {
		return t, true, nil
	}
}

// check and clear node reboot state/status
// fails if expected to reboot, but did not reboot
func (k *Kube) clearNodeReboot() error {
	if rebootTime, rebooting, err := k.checkReboot(); err != nil {
		return err

	} else if !rebooting {
		log.Printf("Initialized kube node %v (not rebooting)", k.node)

		return nil

	} else if !k.hostInfo.BootTime.After(rebootTime) {
		return fmt.Errorf("Kube node %v is still rebooting (reboot=%v >= boot=%v)", k.node, rebootTime, k.hostInfo.BootTime)

	} else if err := k.node.SetCondition(MakeRebootConditionRebooted(k.hostInfo.BootTime)); err != nil {
		log.Printf("Failed to update node %v condition: %v", k.node, err)

		return nil

	} else if err := k.node.ClearAnnotation(KubeRebootAnnotation); err != nil {
		return fmt.Errorf("Failed to clear reboot annotation: %v", err)

	} else {
		log.Printf("Kube node %v was rebooted (reboot=%v < boot=%v)...", k.node, rebootTime, k.hostInfo.BootTime)

		return nil
	}
}

// uncordon if drained before reboot
func (k *Kube) clearNodeDrain() error {
	if changed, err := k.node.SetSchedulableIfAnnotated(KubeDrainAnnotation); err != nil {
		return fmt.Errorf("Failed to clear node drain state: %v", err)
	} else if changed {
		log.Printf("Uncordoned drained kube node %v (with annotation %v)", k.node, KubeDrainAnnotation)
		return nil
	} else {
		log.Printf("Kube node %v is not marked as drained (with annotation %v)", k.node, KubeDrainAnnotation)
		return nil
	}
}

// release lock if still acquired
func (k *Kube) clearLock() error {
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

func (k *Kube) AcquireLock(ctx context.Context) error {
	if k == nil || k.lock == nil {
		log.Printf("Skip kube locking")
		return nil
	}

	log.Printf("Acquiring kube lock...")

	return k.lock.Acquire(ctx)
}

func (k *Kube) ReleaseLock() error {
	if k == nil || k.lock == nil {
		log.Printf("Skip kube unlocking")
		return nil
	}

	log.Printf("Releasing kube lock...")

	return k.lock.Release()
}

// Update node status condition based on function execution
func (k *Kube) UpdateHostStatus(status hosts.Status, upgradeErr error) error {
	if k == nil || k.node == nil {
		log.Printf("Skip updating kube node condition")
		return nil
	}

	log.Printf("Update kube node %v condition for status=%v with error: %v", k.node, status, upgradeErr)

	if err := k.node.SetCondition(
		MakeUpgradeCondition(status, upgradeErr),
		MakeRebootCondition(k.hostInfo, status, upgradeErr),
	); err != nil {
		log.Printf("Failed to update node %v condition: %v", k.node, err)
	}

	return nil
}

func (k *Kube) DrainNode() error {
	if k == nil || k.node == nil {
		return fmt.Errorf("No --kube-node configured")
	}

	log.Printf("Draining kube node %v (with annotation %v)...", k.node, KubeDrainAnnotation)

	if err := k.node.SetAnnotation(KubeDrainAnnotation, "true"); err != nil {
		return fmt.Errorf("Failed to set node annotation for drain: %v", err)
	} else if err := kubectl.Drain(k.options.Node); err != nil {
		return fmt.Errorf("Failed to drain node %v: %v", k.options.Node, err)
	} else {
		return nil
	}
}

func (k *Kube) MarkReboot(rebootTime time.Time) error {
	if k == nil || k.node == nil {
		log.Printf("Skip kube node reboot marking")
		return nil
	}

	log.Printf("Marking kube node %v for reboot (with annotation %v=%v)...", k.node, KubeRebootAnnotation, rebootTime)

	if value, err := json.Marshal(rebootTime); err != nil {
		return fmt.Errorf("Failed to marshal reboot annotation: %v", err)
	} else if err := k.node.SetAnnotation(KubeRebootAnnotation, string(value)); err != nil {
		return fmt.Errorf("Failed to set node annotation for reboot: %v", err)
	} else if err := k.node.SetCondition(MakeRebootConditionRebooting(rebootTime)); err != nil {
		return fmt.Errorf("Failed to set node condition for reboot: %v", err)
	} else {
		return nil
	}
}
