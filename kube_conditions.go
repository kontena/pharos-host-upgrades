package main

import (
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

	if err == nil {
		condition.Status = corev1.ConditionTrue
		condition.Reason = "HostUpgradeDone"
		condition.Message = status.UpgradeLog
	} else {
		condition.Status = corev1.ConditionFalse
		condition.Reason = "HostUpgradeFailed"
		condition.Message = err.Error()
	}

	return condition
}

func MakeRebootCondition(status hosts.Status, info hosts.Info) corev1.NodeCondition {
	var condition = corev1.NodeCondition{
		Type:              RebootConditionType,
		LastHeartbeatTime: metav1.Now(),
	}

	if status.RebootRequired {
		condition.Status = corev1.ConditionTrue
		condition.LastTransitionTime = metav1.NewTime(status.RebootRequiredSince)
		condition.Reason = "RebootRequired"
		condition.Message = status.RebootRequiredMessage
	} else {
		condition.Status = corev1.ConditionFalse
		condition.LastTransitionTime = metav1.NewTime(info.BootTime)
		condition.Reason = "UpToDate"
		condition.Message = status.RebootRequiredMessage
	}

	return condition
}
