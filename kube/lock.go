package kube

import (
	"fmt"
	"log"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

const LockAnnotation = "pharos-host-upgrades.kontena.io/lock"

func MakeLock(kube *Kube) Lock {
	return Lock{
		client:     kube.client,
		namespace:  kube.options.Namespace,
		name:       kube.options.DaemonSet,
		annotation: LockAnnotation,
		value:      kube.options.Node,
	}
}

type Lock struct {
	client     *kubernetes.Clientset
	namespace  string
	name       string
	annotation string
	value      string
	timeout    time.Duration
}

func (lock *Lock) String() string {
	return fmt.Sprintf("daemonsets/%v/%v", lock.namespace, lock.name)
}

// test for lock annotation
func (lock *Lock) test(object runtime.Object) (bool, error) {
	if accessor, err := meta.Accessor(object); err != nil {
		return false, fmt.Errorf("meta.Accessor: %v", err)
	} else if value := accessor.GetAnnotations()[lock.annotation]; value == "" || value == lock.value {
		return true, nil
	} else {
		return false, nil
	}
}

// set lock annotation
func (lock *Lock) set(object runtime.Object) error {
	if accessor, err := meta.Accessor(object); err != nil {
		return fmt.Errorf("meta.Accessor: %v", err)
	} else {
		accessor.GetAnnotations()[lock.annotation] = lock.value
	}

	return nil
}

// clear lock annotation
// fails if not set
func (lock *Lock) clear(object runtime.Object) error {
	if accessor, err := meta.Accessor(object); err != nil {
		return fmt.Errorf("meta.Accessor: %v", err)
	} else if value := accessor.GetAnnotations()[lock.annotation]; value != lock.value {
		return fmt.Errorf("Broken lock: %v, expected %v", value, lock.value)
	} else {
		delete(accessor.GetAnnotations(), lock.annotation)
	}

	return nil
}

// get lock object
func (lock *Lock) get() (runtime.Object, error) {
	if obj, err := lock.client.Apps().DaemonSets(lock.namespace).Get(lock.name, metav1.GetOptions{}); err != nil {
		return nil, err
	} else {
		return obj, err
	}
}

// watch lock object
func (lock *Lock) watch(object runtime.Object) (watch.Interface, error) {
	var listOptions metav1.ListOptions

	if accessor, err := meta.Accessor(object); err != nil {
		return nil, fmt.Errorf("meta.Accessor: %v", err)
	} else {
		listOptions.FieldSelector = fields.OneTermEqualSelector("metadata.name", accessor.GetName()).String()
		listOptions.ResourceVersion = accessor.GetResourceVersion()
	}

	if watcher, err := lock.client.Apps().DaemonSets(lock.namespace).Watch(listOptions); err != nil {
		return nil, err
	} else {
		return watcher, err
	}
}

// update lock object
func (lock *Lock) update(object *runtime.Object) error {
	if ds1, ok := (*object).(*appsv1.DaemonSet); !ok {
		return fmt.Errorf("Invalid object: %T", *object)
	} else if ds2, err := lock.client.Apps().DaemonSets(lock.namespace).Update(ds1); err != nil {
		return err
	} else {
		*object = ds2
	}

	return nil
}

func (lock *Lock) testEvent(event watch.Event) (bool, error) {
	switch event.Type {
	case watch.Modified:
		return lock.test(event.Object)
	default:
		return false, fmt.Errorf("Unexpected event: %v", event.Type)
	}
}

// wait for lock to be free
func (lock *Lock) wait() (runtime.Object, error) {
	if obj, err := lock.get(); err != nil {
		return nil, err
	} else if locked, err := lock.test(obj); err != nil {
		return nil, err
	} else if !locked {
		return obj, nil
	} else if watcher, err := lock.watch(obj); err != nil {
		return obj, err
	} else if ev, err := watch.Until(lock.timeout, watcher, lock.testEvent); err != nil {
		return obj, err
	} else {
		return ev.Object, nil
	}
}

// attempt to acquire lock, assuming it is free
func (lock *Lock) acquire(object runtime.Object) error {
	if err := lock.set(object); err != nil {
		return err
	} else if err := lock.update(&object); err != nil {
		return err
	} else {
		return nil
	}
}

// wait for lock to free and acquire it
func (lock *Lock) Acquire() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if object, err := lock.wait(); err != nil {
			return err
		} else if err := lock.acquire(object); err != nil {
			return err
		} else {
			return nil
		}
	})
}

// attempt to clear lock, assuming it is locked
func (lock *Lock) release(object runtime.Object) error {
	if err := lock.clear(object); err != nil {
		return err
	} else if err := lock.update(&object); err != nil {
		return err
	} else {
		return nil
	}
}

// attempt to release lock, assuming it is set
func (lock *Lock) Release() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if object, err := lock.get(); err != nil {
			return err
		} else if err := lock.release(object); err != nil {
			return err
		} else {
			return nil
		}
	})
}

func (lock *Lock) cleanup() {
	if err := lock.Release(); err != nil {
		log.Printf("Failed to release lock %v: %v", lock, err)
	}
}

func (lock *Lock) With(f func() error) error {
	if err := lock.Acquire(); err != nil {
		return fmt.Errorf("Failed to acquire lock %v: %v", lock, err)
	} else {
		defer lock.cleanup()

		return f()
	}
}
