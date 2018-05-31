package kube

import (
	"context"
	"fmt"
	"log"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

type Lock struct {
	client     appsv1client.AppsV1Interface
	namespace  string
	name       string
	annotation string
	value      string
}

func (lock *Lock) String() string {
	return fmt.Sprintf("%v/daemonsets/%v", lock.namespace, lock.name)
}

func (lock *Lock) connect(config *rest.Config) error {
	if client, err := appsv1client.NewForConfig(config); err != nil {
		return err
	} else {
		lock.client = client
	}

	return nil
}

// test for lock annotation
func (lock *Lock) test(object runtime.Object) (value string, available bool, acquired bool) {
	if accessor, err := meta.Accessor(object); err != nil {
		panic(err)
	} else if value := accessor.GetAnnotations()[lock.annotation]; value == "" {
		log.Printf("kube/lock %v: test %v=%v: free", lock, lock.annotation, value)

		return value, true, false

	} else if value == lock.value {
		log.Printf("kube/lock %v: test %v=%v: acquired", lock, lock.annotation, value)

		return value, true, true

	} else {
		log.Printf("kube/lock %v: test %v=%v: locked", lock, lock.annotation, value)

		return value, false, false
	}
}

// set lock annotation
func (lock *Lock) set(object runtime.Object) error {
	if accessor, err := meta.Accessor(object); err != nil {
		panic(err)
	} else {
		log.Printf("kube/lock %v: set %v=%v", lock, lock.annotation, lock.value)

		accessor.GetAnnotations()[lock.annotation] = lock.value
	}

	return nil
}

// clear lock annotation
// fails if not set
func (lock *Lock) clear(object runtime.Object) error {
	if accessor, err := meta.Accessor(object); err != nil {
		panic(err)
	} else if value := accessor.GetAnnotations()[lock.annotation]; value != lock.value {
		return fmt.Errorf("Broken lock: %v=%v, expected %v", lock.annotation, value, lock.value)
	} else {
		log.Printf("kube/lock %v: clear %v=%v", lock, lock.annotation, value)

		delete(accessor.GetAnnotations(), lock.annotation)
	}

	return nil
}

// get lock object
func (lock *Lock) get() (runtime.Object, error) {
	log.Printf("kube/lock %v: get", lock)

	if obj, err := lock.client.DaemonSets(lock.namespace).Get(lock.name, metav1.GetOptions{}); err != nil {
		return nil, fmt.Errorf("Get: %v", err)
	} else {
		return obj, nil
	}
}

// Test lock object
func (lock *Lock) Test() (value string, acquired bool, err error) {
	if object, err := lock.get(); err != nil {
		return "", false, err
	} else {
		value, _, acquired := lock.test(object)

		return value, acquired, nil
	}
}

// watch lock object
func (lock *Lock) watch(object runtime.Object) (watch.Interface, error) {
	var listOptions metav1.ListOptions

	if accessor, err := meta.Accessor(object); err != nil {
		panic(err)
	} else {
		listOptions.FieldSelector = fields.OneTermEqualSelector("metadata.name", accessor.GetName()).String()
		listOptions.ResourceVersion = accessor.GetResourceVersion()
	}

	log.Printf("kube/lock %v: watch %#v", lock, listOptions)

	if watcher, err := lock.client.DaemonSets(lock.namespace).Watch(listOptions); err != nil {
		return nil, fmt.Errorf("Watch: %v", err)
	} else {
		return watcher, nil
	}
}

// update lock object
func (lock *Lock) update(object *runtime.Object) error {
	log.Printf("kube/lock %v: update", lock)

	if ds1, ok := (*object).(*appsv1.DaemonSet); !ok {
		panic(fmt.Errorf("Invalid object: %T", *object))
	} else if ds2, err := lock.client.DaemonSets(lock.namespace).Update(ds1); err != nil {
		return err // unmodified for RetryOnConflict
	} else {
		*object = ds2
	}

	return nil
}

func (lock *Lock) testEvent(event watch.Event) (bool, error) {
	switch event.Type {
	case watch.Modified:
		if _, available, _ := lock.test(event.Object); available {
			return true, nil
		}
	default:
		return false, fmt.Errorf("Unexpected event: %v", event.Type)
	}

	return false, nil
}

func contextTimeout(ctx context.Context) time.Duration {
	if deadline, ok := ctx.Deadline(); !ok {
		return time.Duration(0)
	} else {
		return deadline.Sub(time.Now())
	}
}

// wait for lock to be free
func (lock *Lock) wait(ctx context.Context) (runtime.Object, error) {
	log.Printf("kube/lock %v: wait", lock)

	if obj, err := lock.get(); err != nil {
		return obj, err
	} else if _, available, _ := lock.test(obj); available {
		// fastpath
		return obj, nil
	} else if watcher, err := lock.watch(obj); err != nil {
		return obj, err
	} else if ev, err := watch.Until(contextTimeout(ctx), watcher, lock.testEvent); err != nil {
		log.Printf("kube/lock %v: wait err: %v", lock, err)
		return obj, err
	} else {
		log.Printf("kube/lock %v: wait ok", lock)
		return ev.Object, nil
	}
}

// attempt to acquire lock, assuming it is free
func (lock *Lock) acquire(object runtime.Object) error {
	log.Printf("kube/lock %v: acquire", lock)

	if err := lock.set(object); err != nil {
		return err
	} else if err := lock.update(&object); err != nil {
		return err
	} else {
		return nil
	}
}

// wait for lock to free and acquire it
func (lock *Lock) Acquire(ctx context.Context) error {
	// XXX: watch timeout errors get lost: https://github.com/kubernetes/client-go/issues/427
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := ctx.Err(); err != nil {
			return err
		} else if object, err := lock.wait(ctx); err != nil {
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
	log.Printf("kube/lock %v: release", lock)

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
