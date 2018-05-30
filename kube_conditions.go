package main

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kontena/pharos-host-upgrades/hosts"
)

const UpgradeConditionType corev1.NodeConditionType = "HostUpgrades"
const RebootConditionType corev1.NodeConditionType = "HostUpgradesReboot"

func MakeUpgradeCondition(status hosts.Status, err error) corev1.NodeCondition {
	var condition = corev1.NodeCondition{
		Type:               UpgradeConditionType,
		LastHeartbeatTime:  metav1.Now(),
		LastTransitionTime: metav1.Now(), // only on changes?
	}

	if err != nil {
		condition.Status = corev1.ConditionUnknown
		condition.Reason = "UpgradeFailed"
		condition.Message = err.Error()
	} else if status.RebootRequired {
		condition.Status = corev1.ConditionFalse
		condition.Reason = "RebootRequired"
		condition.Message = status.UpgradeLog
	} else {
		condition.Status = corev1.ConditionTrue
		condition.Reason = "UpToDate"
		condition.Message = status.UpgradeLog
	}

	return condition
}

func MakeRebootCondition(info hosts.Info, status hosts.Status, upgradeErr error) corev1.NodeCondition {
	var condition = corev1.NodeCondition{
		Type:              RebootConditionType,
		LastHeartbeatTime: metav1.Now(),
	}

	if status.RebootRequired {
		condition.Status = corev1.ConditionTrue
		condition.LastTransitionTime = metav1.NewTime(status.RebootRequiredSince)
		condition.Reason = "RebootRequired"
		condition.Message = status.RebootRequiredMessage
	} else if upgradeErr != nil {
		condition.Status = corev1.ConditionUnknown
		condition.LastTransitionTime = metav1.Now()
		condition.Reason = "UpgradeFailed"
	} else {
		condition.Status = corev1.ConditionFalse
		condition.LastTransitionTime = metav1.NewTime(info.BootTime)
		condition.Reason = "UpToDate"
		condition.Message = status.RebootRequiredMessage
	}

	return condition
}

func MakeRebootConditionRebooting(rebootTime time.Time) corev1.NodeCondition {
	var condition = corev1.NodeCondition{
		Type:              RebootConditionType,
		LastHeartbeatTime: metav1.Now(),
	}

	condition.Status = corev1.ConditionTrue
	condition.LastTransitionTime = metav1.NewTime(rebootTime)
	condition.Reason = "Rebooting"

	return condition
}

func MakeRebootConditionRebooted(bootTime time.Time) corev1.NodeCondition {
	var condition = corev1.NodeCondition{
		Type:              RebootConditionType,
		LastHeartbeatTime: metav1.Now(),
	}

	condition.Status = corev1.ConditionFalse
	condition.LastTransitionTime = metav1.NewTime(bootTime)
	condition.Reason = "Rebooted"

	return condition
}
