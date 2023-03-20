/*
Copyright © 2023 The vm-crd-create Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	"vm-crd-create/pkg/apis/cloudnative/v1alpha1"
	"vm-crd-create/pkg/driver"
	cloudnative "vm-crd-create/pkg/generated/clientset/versioned"
	cloudnativeinformer "vm-crd-create/pkg/generated/informers/externalversions"
	listerv1alpha1 "vm-crd-create/pkg/generated/listers/cloudnative/v1alpha1"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const (
	resouceName   = "VirtualMachine"
	periodSec     = 20
	finalizerName = "finalizer.cloudnative.lich"
)

type Controller struct {
	clientset cloudnative.Interface
	informer  cloudnativeinformer.SharedInformerFactory
	lister    listerv1alpha1.VirtualMachineLister
	synced    cache.InformerSynced
	queue     workqueue.RateLimitingInterface
	vm        driver.Interface
}

func New(clientset cloudnative.Interface, informer cloudnativeinformer.SharedInformerFactory, vm driver.Interface) *Controller {
	vmInformer := informer.Cloudnative().V1alpha1().VirtualMachines()
	controller := &Controller{
		clientset: clientset,
		informer:  informer,
		vm:        vm,
		lister:    vmInformer.Lister(),
		synced:    vmInformer.Informer().HasSynced,
		queue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), resouceName),
	}

	vmInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueue,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueue(new)
		},
		// DeleteFunc: controller.deleteObject,
	})
	return controller
}

func (c *Controller) Run(stopCh <-chan struct{}, threadiness int) error {
	defer utilruntime.HandleCrash()
	defer c.Stop()

	klog.Info("Starting the controller")
	klog.Info("Waiting for the informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.synced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

func (c *Controller) Stop() {
	klog.Info("Stopping the controller")
	c.queue.ShutDown()
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.queue.Done(obj)
		key, ok := obj.(string)
		if !ok {
			c.queue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("Controller expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.syncHandler(key); err != nil {
			c.queue.AddRateLimited(key)
			return fmt.Errorf("Controller error syncing '%s': %s, requeuing", key, err.Error())
		}

		c.queue.Forget(obj)
		glog.Infof("Controller successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (c *Controller) enqueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.queue.Add(key)
}

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return err
	}

	vm, err := c.lister.VirtualMachines(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("virtualmachine '%s' in work queue no longer exists", key))
			return err
		}
		return err
	}

	switch vm.Status.Phase {
	case v1alpha1.VirtualMachineNone:
		if err := c.makeCreatingPhase(vm); err != nil {
			return err
		}

	case v1alpha1.VirtualMachineCreating, v1alpha1.VirtualMachineFailed:
		if err := c.createServer(vm); err != nil {
			return err
		}
	case v1alpha1.VirtualMachineActive:
		if !vm.ObjectMeta.DeletionTimestamp.IsZero() {
			if err := c.makeTerminatingPhase(vm); err != nil {
				return err
			}
			return nil
		}

		if err := c.updateUsage(vm); err != nil {
			return err
		}
	case v1alpha1.VirtualMachineTerminating:
		if err := c.deleteServer(vm); err != nil {
			return err
		}
	}

	return nil
}

// func (c *Controller) deleteObject(obj interface{}) {
// 	vm := obj.(*v1alpha1.VirtualMachine)
// 	if err := c.vm.DeleteServer(vm.Name); err != nil {
// 		klog.Errorf("failed to delete the '%s' server: %v", vm.Name, err)
// 	}
// }

func (c *Controller) makeCreatingPhase(vm *v1alpha1.VirtualMachine) error {
	vmCopy := vm.DeepCopy()
	klog.Info("makeCreatingPhase")
	return c.updateVM(vmCopy, v1alpha1.VirtualMachineCreating, nil)
}

func (c *Controller) makeTerminatingPhase(vm *v1alpha1.VirtualMachine) error {
	vmCopy := vm.DeepCopy()
	klog.Info("makeTerminatingPhase")
	return c.updateVM(vmCopy, v1alpha1.VirtualMachineTerminating, nil)
}

func (c *Controller) createServer(vm *v1alpha1.VirtualMachine) error {
	vmCopy := vm.DeepCopy()
	klog.Info("createServer")
	ok, _ := c.vm.IsServerExist(vm.Name)
	if !ok {
		klog.Info("the cpu is: ", vm.Spec.Cpu.Value())
		klog.Info("the memory is: ", vm.Spec.Memory.Value())
		req := &driver.CreateRequest{
			Name:   vm.Name,
			CPU:    vm.Spec.Cpu.Value(),
			Memory: vm.Spec.Memory.Value(),
		}
		resp, err := c.vm.CreateServer(req)
		if err != nil {
			if err := c.updateVM(vmCopy, v1alpha1.VirtualMachineFailed, err); err != nil {
				return err
			}
			return err
		}
		vmCopy.Status.Server.ID = resp.ID

		if err := c.appendServerStatus(vmCopy); err != nil {
			return err
		}

		addFinalizer(&vmCopy.ObjectMeta, finalizerName)
		if err := c.updateVM(vmCopy, v1alpha1.VirtualMachineActive, nil); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) appendServerStatus(vm *v1alpha1.VirtualMachine) error {
	status, err := c.vm.GetServerStatus(vm.Name)
	if err != nil {
		return err
	}

	vm.Status.Server.Usage.CPU = status.CPUPercentage
	vm.Status.Server.Usage.Memory = status.MemoryPercentage
	vm.Status.Server.State = status.State
	return nil
}

func (c *Controller) updateUsage(vm *v1alpha1.VirtualMachine) error {
	vmCopy := vm.DeepCopy()
	t := subtractTime(vmCopy.Status.LastUpdateTime.Time)
	if t.Seconds() > periodSec {
		if err := c.appendServerStatus(vmCopy); err != nil {
			return err
		}

		if err := c.updateVM(vmCopy, v1alpha1.VirtualMachineActive, nil); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) updateVM(vm *v1alpha1.VirtualMachine, phase v1alpha1.VirtualMachinePhase, reason error) error {
	vm.Status.Reason = ""
	if reason != nil {
		vm.Status.Reason = reason.Error()
	}

	updateTime := metav1.NewTime(time.Now())
	vm.Status.Phase = phase
	vm.Status.LastUpdateTime = updateTime
	_, err := c.clientset.CloudnativeV1alpha1().VirtualMachines(vm.Namespace).Update(context.TODO(), vm, metav1.UpdateOptions{})
	return err
}

func (c *Controller) deleteServer(vm *v1alpha1.VirtualMachine) error {
	vmCopy := vm.DeepCopy()
	if err := c.vm.DeleteServer(vmCopy.Name); err != nil {
		// Requeuing object to workqueue for retrying
		return err
	}

	removeFinalizer(&vmCopy.ObjectMeta, finalizerName)
	if err := c.updateVM(vmCopy, v1alpha1.VirtualMachineTerminating, nil); err != nil {
		return err
	}
	return nil
}
