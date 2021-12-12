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

package tunnel

import (
	"context"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"time"

	corev1 "github.com/cokeos/zero/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const UpdatePeriod = time.Second * 10

// TunnelReconciler reconciles a Tunnel object
type TunnelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.cokeos.io,resources=tunnels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.cokeos.io,resources=tunnels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.cokeos.io,resources=tunnels/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the Tunnel object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *TunnelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var (
		tunnel  = &corev1.Tunnel{}
		service = &v1.Service{}
	)
	tunnelErr := r.Get(ctx, req.NamespacedName, tunnel)
	serviceErr := r.Get(ctx, req.NamespacedName, service)

	if tunnelErr != nil {
		if apierrors.IsNotFound(tunnelErr) {
			if err := r.Delete(ctx, service); err != nil {
				klog.Error(err)
			}
		} else {
			klog.Error(tunnelErr)
		}
		return ctrl.Result{}, nil
	}

	if tunnel.DeletionTimestamp != nil {
		if err := r.Delete(ctx, service); err != nil {
			klog.Error(err)
		}
		return ctrl.Result{}, nil
	}

	if serviceErr != nil {
		if apierrors.IsNotFound(tunnelErr) {
			service = generateService(tunnel)
			return ctrl.Result{}, r.Create(ctx, service)
		} else {
			klog.Error(serviceErr)
		}
	}

	return ctrl.Result{}, nil
}

func (r *TunnelReconciler) SyncService() {
	var (
		ctx  = context.TODO()
		list = &v1.ServiceList{}
	)
	err := r.List(ctx, list, &client.ListOptions{LabelSelector: labels.SelectorFromSet(map[string]string{
		LabelKey: LabelValue,
	})})
	if err != nil {
		klog.Errorf("List Services Error: %v", err)
		return
	}
	for _, svc := range list.Items {
		tunnel := &corev1.Tunnel{}
		err = r.Get(ctx, types.NamespacedName{
			Name:      svc.GetName(),
			Namespace: svc.GetNamespace(),
		}, tunnel)
		if err != nil {
			klog.Errorf("Get Tunnel Error: %v", err)
			continue
		}
		tunnel.Status.Conditions = svc.Status.Conditions
		err = r.Status().Update(ctx, tunnel)
		if err != nil {
			klog.Errorf("Update Tunnel Conditions Error: %v", err)
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *TunnelReconciler) SetupWithManager(mgr ctrl.Manager, ctx context.Context) error {
	go func() {
		if mgr.GetCache().WaitForCacheSync(ctx) {
			go wait.Until(r.SyncService, UpdatePeriod, ctx.Done())
		}
	}()
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Tunnel{}).
		Complete(r)
}
