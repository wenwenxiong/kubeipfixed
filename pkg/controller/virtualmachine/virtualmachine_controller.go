/*
Copyright 2019 The KubeMacPool Authors.

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

package virtualmachine

import (
	"context"
	"fmt"
	"math/rand"

	"k8s.io/apimachinery/pkg/runtime"
	kubevirt "kubevirt.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	ip_manager "github.com/wenwenxiong/kubeipfixed/pkg/ip-manager"
)

var log = logf.Log.WithName("VirtualMachine Controller")

// Add creates a new Policy Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, poolManager *ip_manager.IPManager) error {
	if poolManager.IsKubevirtEnabled() {
		return add(mgr, newReconciler(mgr, poolManager))
	}

	return nil
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, poolManager *ip_manager.IPManager) reconcile.Reconciler {
	return &ReconcilePolicy{Client: mgr.GetClient(), scheme: mgr.GetScheme(), poolManager: poolManager}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("virtualmachine-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Pod
	err = c.Watch(&source.Kind{Type: &kubevirt.VirtualMachine{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcilePolicy{}

// ReconcilePolicy reconciles a Policy object
type ReconcilePolicy struct {
	client.Client
	scheme      *runtime.Scheme
	poolManager *ip_manager.IPManager
}

// Reconcile reads that state of the cluster for a virtual machine object and makes changes based on the state
func (r *ReconcilePolicy) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	//used for multi thread log separation
	reconcileRequestId := rand.Intn(100000)
	logger := log.WithName("Reconcile").WithValues("RequestId", reconcileRequestId, "vmFullName", fmt.Sprintf("vm/%s/%s", request.Namespace, request.Name))
	logger.Info("got a virtual machine event in the controller")

	instance := &kubevirt.VirtualMachine{}
	err := r.Get(ctx, request.NamespacedName, instance)

	return reconcile.Result{}, err
}
