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

package unit

import (
	"context"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	corev1 "github.com/cokeos/zero/api/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// UnitReconciler reconciles a Unit object
type UnitReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.cokeos.io,resources=units,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.cokeos.io,resources=units/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.cokeos.io,resources=units/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *UnitReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var (
		pod  = &v1.Pod{}
		unit = &corev1.Unit{}
	)

	// Unit 查询
	unitErr := r.Get(ctx, req.NamespacedName, unit)
	podErr := r.Get(ctx, req.NamespacedName, pod)

	// 删除逻辑
	if unitErr != nil {
		if apierrors.IsNotFound(unitErr) {
			if err := r.Delete(ctx, pod); err != nil {
				klog.Error(err)
			}
		} else {
			klog.Error(unitErr)
		}
		return ctrl.Result{}, nil
	}
	if unit.DeletionTimestamp != nil {
		if err := r.Delete(ctx, pod); err != nil {
			klog.Error(err)
		}
		return ctrl.Result{}, nil
	}

	// 创建逻辑
	if podErr != nil {
		if apierrors.IsNotFound(podErr) {
			pod = generatePod(unit)
			return ctrl.Result{}, r.Create(ctx, pod)
		} else {
			klog.Error(podErr)
		}
	}

	return ctrl.Result{}, nil
}

const (
	UpdatePeriod = time.Second * 10
)

// SetupWithManager sets up the controller with the Manager.
func (r *UnitReconciler) SetupWithManager(mgr ctrl.Manager, ctx context.Context) error {
	go func() {
		if mgr.GetCache().WaitForCacheSync(ctx) {
			go wait.Until(r.SyncPods, UpdatePeriod, ctx.Done())
		}
	}()
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Unit{}).
		Complete(r)
}

func (r *UnitReconciler) SyncPods() {
	var (
		ctx  = context.TODO()
		list = &v1.PodList{}
	)
	err := r.List(ctx, list, &client.ListOptions{LabelSelector: labels.SelectorFromSet(map[string]string{
		LabelKey: LabelValue,
	})})
	if err != nil {
		klog.Errorf("List Pods Error: %v", err)
		return
	}
	for _, pod := range list.Items {
		unit := &corev1.Unit{}
		err = r.Get(ctx, types.NamespacedName{
			Name:      pod.GetName(),
			Namespace: pod.GetNamespace(),
		}, unit)
		if err != nil {
			klog.Errorf("Get Unit Error: %v", err)
			continue
		}
		unit.Status.Phase = pod.Status.Phase
		err = r.Status().Update(ctx, unit)
		if err != nil {
			klog.Errorf("Update Unit Phase Error: %v", err)
		}
	}
}
