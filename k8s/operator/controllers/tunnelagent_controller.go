/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"

	//appsv1 "k8s.io/api/apps/v1"
	//corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tunnelv1 "github.com/sebgrz/operator/api/v1"
)

// TunnelAgentReconciler reconciles a TunnelAgent object
type TunnelAgentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tunnel.my.domain,resources=tunnelagents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tunnel.my.domain,resources=tunnelagents/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tunnel.my.domain,resources=tunnelagents/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TunnelAgent object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *TunnelAgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info(fmt.Sprintf("tunnelAgent req: %+v", req))

	// Should check 2 types resources - TunnelAgent and Service
	// If TunnelAgent was created:
	// - check if Agent deployment is created (that means - conenction Service <-> Agent is running)
	// - if not - check if service with required label exists (tunnel-agent) and the label has value equal with TunnalAgent resource name
	// 	- if yes - create deployment with data from CRD and adress of service
	// If TunnelAgent is deleted - destroy deployment if exists
	// TODO: additional condition for update event
	var tunnelAgent tunnelv1.TunnelAgent
	// existingDeployment := &appsv1.Deployment{}
	// existingService := &corev1.Service{}

	err := r.Get(ctx, req.NamespacedName, &tunnelAgent)
	if err != nil {
		if errors.IsNotFound(err) {

		} else {
			logger.Error(err, "check TunnelAgent object")
		}
	}
	logger.Info(fmt.Sprintf("TunnelAgent %+v", tunnelAgent))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TunnelAgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tunnelv1.TunnelAgent{}).
		Complete(r)
}
