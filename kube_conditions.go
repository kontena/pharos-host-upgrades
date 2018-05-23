package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const UpgradeConditionType corev1.NodeConditionType = "HostUpgrades"

func MakeUpgradeCondition(err error) corev1.NodeCondition {
	var condition = corev1.NodeCondition{
		Type:               UpgradeConditionType,
		LastHeartbeatTime:  metav1.Now(),
		LastTransitionTime: metav1.Now(), // XXX: only on changes?
	}

	if err == nil {
		condition.Status = corev1.ConditionTrue
		condition.Reason = "HostUpgradeDone"
	} else {
		condition.Status = corev1.ConditionFalse
		condition.Reason = "HostUpgradeFailed"
		condition.Message = err.Error()
	}

	return condition
}
