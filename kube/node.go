package kube

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

type Node struct {
	client corev1client.CoreV1Interface
	name   string
}

func (node *Node) String() string {
	return fmt.Sprintf("%v", node.name)
}

func (node *Node) connect(config *rest.Config) error {
	if client, err := corev1client.NewForConfig(config); err != nil {
		return err
	} else {
		node.client = client
	}

	return nil
}

func (node *Node) get() (*corev1.Node, error) {
	if obj, err := node.client.Nodes().Get(node.name, metav1.GetOptions{}); err != nil {
		return nil, fmt.Errorf("Get %v: %v", node, err)
	} else {
		return obj, nil
	}
}

func (node *Node) setCondition(obj *corev1.Node, condition corev1.NodeCondition) {
	for i, c := range obj.Status.Conditions {
		if c.Type == condition.Type {
			obj.Status.Conditions[i] = condition
			return
		}
	}

	obj.Status.Conditions = append(obj.Status.Conditions, condition)
}

func (node *Node) setConditions(obj *corev1.Node, conditions []corev1.NodeCondition) {
	for _, condition := range conditions {
		node.setCondition(obj, condition)
	}
}

func (node *Node) getCondition(obj *corev1.Node, conditionType corev1.NodeConditionType) (condition corev1.NodeCondition, exists bool) {
	for _, c := range obj.Status.Conditions {
		if c.Type == conditionType {
			return c, true
		}
	}

	return condition, false
}

func (node *Node) GetCondition(conditionType corev1.NodeConditionType) (condition corev1.NodeCondition, exists bool, err error) {
	if obj, err := node.get(); err != nil {
		return condition, false, err
	} else if condition, exists := node.getCondition(obj, conditionType); !exists {
		return condition, false, nil
	} else {
		return condition, true, nil
	}
}

func (node *Node) HasCondition(conditionType corev1.NodeConditionType) (bool, error) {
	if _, exists, err := node.GetCondition(conditionType); err != nil {
		return exists, err
	} else {
		return exists, nil
	}
}

func (node *Node) SetCondition(conditions ...corev1.NodeCondition) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if obj, err := node.get(); err != nil {
			return err
		} else if node.setConditions(obj, conditions); err != nil { // XXX: control flow hack, can't fail
			return err
		} else if _, err := node.client.Nodes().UpdateStatus(obj); err != nil {
			return err // unmodified for RetryOnConflict
		} else {
			return nil
		}
	})
}

func (node *Node) InitCondition(conditionType corev1.NodeConditionType) error {
	return node.SetCondition(corev1.NodeCondition{
		Type:              conditionType,
		Status:            corev1.ConditionUnknown,
		LastHeartbeatTime: metav1.Now(),
	})
}
