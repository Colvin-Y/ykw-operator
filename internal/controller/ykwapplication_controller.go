/*
Copyright 2023 Colvin-Y.

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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ykwapiv1 "github.com/Colvin-Y/ykw-operator/api/v1"
)

// YkwApplicationReconciler reconciles a YkwApplication object
type YkwApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ykwapi.cn.hrimfaxi,resources=ykwapplications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ykwapi.cn.hrimfaxi,resources=ykwapplications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ykwapi.cn.hrimfaxi,resources=ykwapplications/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the YkwApplication object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *YkwApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	// get YkwApplication
	app := &ykwapiv1.YkwApplication{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		if errors.IsNotFound(err) {
			l.Info("this resource(YkwApplication) is not found")

			// 这个资源被删了，那么就可以不再关注它了
			return ctrl.Result{}, nil
		}
		l.Error(err, "failed to get YkwApplication")
		// 20s 一个轮回
		return ctrl.Result{RequeueAfter: 20 * time.Second}, err
	}

	// create pods
	for i := 0; i < int(app.Spec.Replicas); i++ {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        fmt.Sprintf("%s-%d", app.Name, i),
				Namespace:   app.Namespace,
				Labels:      app.Labels,
				Annotations: app.Annotations,
			},
			Spec: app.Spec.Template.Spec,
		}
		if err := r.Create(ctx, pod); err != nil {
			l.Error(err, fmt.Sprintf("failed to create pod[%s]", pod.Name))
			return ctrl.Result{RequeueAfter: 20 * time.Second}, err
		}
		l.Info(fmt.Sprintf("create pod[%s] success", pod.Name))
	}
	l.Info("all pods has created")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *YkwApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ykwapiv1.YkwApplication{}).
		Complete(r)
}
