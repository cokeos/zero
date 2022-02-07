/*
Copyright 2021.

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

package tiny

import (
	"context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"sync"
	"time"

	corev1 "github.com/cokeos/zero/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TinyReconciler reconciles a Tiny object
type TinyReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// PortMap
	PortMap map[int32]bool
	Mu      sync.RWMutex
}

//+kubebuilder:rbac:groups=core.cokeos.io,resources=tinies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.cokeos.io,resources=tinies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.cokeos.io,resources=tinies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the Tiny object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *TinyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var (
		tiny      = &corev1.Tiny{}
		unit      = &corev1.Unit{}
		tunnel    = &corev1.Tunnel{}
		tinyErr   = r.Get(ctx, req.NamespacedName, tiny)
		unitErr   = r.Get(ctx, req.NamespacedName, unit)
		tunnelErr = r.Get(ctx, req.NamespacedName, tunnel)
	)

	if tinyErr != nil {
		if apierrors.IsNotFound(tinyErr) {
			if err := r.Delete(ctx, unit); err != nil {
				klog.Error(err)
			}
			if err := r.Delete(ctx, tunnel); err != nil {
				klog.Error(err)
			}
		} else {
			klog.Error(tinyErr)
		}
		return ctrl.Result{}, nil
	}

	if tiny.DeletionTimestamp != nil {
		if err := r.Delete(ctx, unit); err != nil {
			klog.Error(err)
		}
		if err := r.Delete(ctx, tunnel); err != nil {
			klog.Error(err)
		}
		return ctrl.Result{}, nil
	}

	if unitErr != nil {
		if apierrors.IsNotFound(unitErr) {
			unit = generateUnit(tiny)
			if err := r.Create(ctx, unit); err != nil {
				klog.Error(err)
			}
		}
	}

	if tunnelErr != nil {
		if apierrors.IsNotFound(tunnelErr) {
			port := r.FindSSHAvailablePort()
			tunnel = generateTunnel(tiny, port)
			r.AddUsedPort(port)
			if err := r.Create(ctx, tunnel); err != nil {
				klog.Error(err)
			}
			tiny.Status.NodePort = port
			if err := r.Status().Update(ctx, tiny); err != nil {
				klog.Error(err)
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TinyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.InitNodeMap()
	go wait.Forever(r.UpdatePortMap, time.Minute)
	go wait.Forever(r.SyncTiny, time.Minute)
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Tiny{}).
		Complete(r)
}

func (r *TinyReconciler) InitNodeMap() {
	r.PortMap = make(map[int32]bool, 0)
	r.Mu = sync.RWMutex{}
	for i := 30000; i < 32000; i++ {
		r.PortMap[int32(i)] = false
	}
}

func (r *TinyReconciler) UpdatePortMap() {
	tunnelList := &corev1.TunnelList{}
	err := r.Client.List(context.TODO(), tunnelList)
	if err != nil {
		klog.Error(err)
		return
	}
	for _, tunnel := range tunnelList.Items {
		for _, port := range tunnel.Spec.Ports {
			r.AddUsedPort(port.NodePort)
		}
	}
}

func (r *TinyReconciler) AddUsedPort(port int32) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.PortMap[port] = true
}

func (r *TinyReconciler) FindSSHAvailablePort() int32 {
	r.Mu.RLock()
	defer r.Mu.RUnlock()
	for i := 30000; i < 32000; i++ {
		if !r.PortMap[int32(i)] {
			return int32(i)
		}
	}
	return -1
}

func (r *TinyReconciler) SyncTiny() {
	var (
		ctx  = context.TODO()
		list = &corev1.TinyList{}
	)
	err := r.List(ctx, list)
	if err != nil {
		klog.Errorf("List Unit Error: %v", err)
		return
	}
	for _, tiny := range list.Items {
		unit := &corev1.Unit{}
		err = r.Get(ctx, types.NamespacedName{
			Name:      tiny.GetName(),
			Namespace: tiny.GetNamespace(),
		}, unit)
		if err != nil {
			klog.Errorf("Get Unit Error: %v", err)
			continue
		}

		if tiny.Spec.Days != 0 && tiny.CreationTimestamp.Add(
			time.Duration(tiny.Spec.Days*24)*time.Hour).Before(time.Now()) {
			err = r.Delete(ctx, tiny.DeepCopy())
			if err != nil {
				klog.Errorf("Delete Tiny Error: %v", err)
			}
		}

		tiny.Status.Phase = unit.Status.Phase
		err = r.Status().Update(ctx, tiny.DeepCopy())
		if err != nil {
			klog.Errorf("Update Tiny Phase Error: %v", err)
		}
	}

}
