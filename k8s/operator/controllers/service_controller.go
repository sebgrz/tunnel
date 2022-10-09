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
	"time"

	tunnelv1 "github.com/sebgrz/operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/strings/slices"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const serviceFinalizer = "example.com/service-tunnel-agent-finalizer"

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tunnel.my.domain,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tunnel.my.domain,resources=services/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tunnel.my.domain,resources=services/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Service object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// - is not necessary to check Agent deployment exists because "created" state means Service is new
	// - check if TunnelAgent resource exists with a name as "tunnel-agent" label value
	// If Service is deleted - destroy deployment
	// TODO: case for update
	var service corev1.Service
	err := r.Get(ctx, req.NamespacedName, &service)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("Service object not found - deleted %+v", service))
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "get Service object")
			return ctrl.Result{}, err
		}
	}

	logger.Info(fmt.Sprintf("service: %+v", service))

	result, err := r.deleteFinalizerIfPossible(ctx, &service, r.deleteDeployment)
	if err != nil || result != nil{
		return *result, err
	}

	if err = r.addFinalizer(ctx, &service); err != nil {
		return ctrl.Result{}, err
	}

	// Get TunnelAgent object
	tunnelAgentName := service.Labels["tunnel-agent"]
	servicePort := service.Spec.Ports[0].Port
	destination := fmt.Sprintf("%s:%d", service.Name, servicePort)

	var tunnelAgent tunnelv1.TunnelAgent
	err = r.Get(ctx, types.NamespacedName{Name: tunnelAgentName, Namespace: service.Namespace}, &tunnelAgent)
	if err != nil {
		logger.Error(err, "get TunnelAgent object")
		return ctrl.Result{}, nil
	}

	deploymentConfig := tunnelAgentDeploymentConfig{
		Name:        tunnelAgentName,
		Namespace:   service.Namespace,
		Destination: destination,
		Resource:    &tunnelAgent,
	}
	tunnelAgentDeployment := createTunnelAgentDeployment(deploymentConfig)
	logger.Info("creating new TunnelAgent Deployment")

	err = r.Create(ctx, tunnelAgentDeployment)
	if err != nil {
		logger.Error(err, "TunnelAgent Deployment failed")
	}

	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) deleteDeployment(ctx context.Context, service *corev1.Service) error {
	logger := log.FromContext(ctx)
	tunnelAgentName := service.Labels["tunnel-agent"]

	var deployment appsv1.Deployment
	err := r.Get(ctx, types.NamespacedName{Name: deploymentName(tunnelAgentName), Namespace: service.Namespace}, &deployment)
	if err != nil {
		logger.Error(err, "get TunnelAgent deployment object while deletion process")
		return nil
	}
	err = r.Delete(ctx, &deployment)
	if err != nil {
		logger.Error(err, "delete TunnelAgent deployment object")
		return err
	}
	return nil
}

func (r *ServiceReconciler) deleteFinalizerIfPossible(ctx context.Context, service *corev1.Service, whenDeletedFunc func(context.Context, *corev1.Service) error) (*ctrl.Result, error) {
	logger := log.FromContext(ctx)
	// Check if the Service instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isServiceMarkedToDelete := service.GetDeletionTimestamp() != nil
	if isServiceMarkedToDelete {
		if controllerutil.ContainsFinalizer(service, serviceFinalizer) {
			// delete logic here
			if err := whenDeletedFunc(ctx, service); err != nil {
				logger.Error(err, "whenDeletedFunc logic failed")
				return &ctrl.Result{Requeue: true, RequeueAfter: time.Minute * 1}, err
			}

			// Remove serviceFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(service, serviceFinalizer)
			if err := r.Update(ctx, service); err != nil {
				logger.Error(err, "remove service finalizer failed")
				return &ctrl.Result{}, err
			}
		}

		return &ctrl.Result{}, nil
	}

	return nil, nil
}

func (r *ServiceReconciler) addFinalizer(ctx context.Context, service *corev1.Service) error {
	logger := log.FromContext(ctx)

	if !controllerutil.ContainsFinalizer(service, serviceFinalizer) {
		controllerutil.AddFinalizer(service, serviceFinalizer)
		err := r.Update(ctx, service)
		if err != nil {
			logger.Error(err, "finalizer cache")
			return err
		}
		logger.Info("finalizer added to service: " + service.Name)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	filterByLabelFunc := func(obj client.Object) bool {
		if label, exists := obj.GetLabels()["tunnel-agent"]; exists {
			mgr.GetLogger().Info("event label: " + label)
			return true
		}

		return false
	}
	filterPredicate := predicate.Funcs{
		CreateFunc: func(ce event.CreateEvent) bool {
			return filterByLabelFunc(ce.Object)
		},
		DeleteFunc: func(de event.DeleteEvent) bool {
			return filterByLabelFunc(de.Object)
		},
		UpdateFunc: func(ue event.UpdateEvent) bool {
			// TODO: check if Service has lost the label

			// If update was about to add finalizer - ignore that event!
			return !(!slices.Contains(ue.ObjectOld.GetFinalizers(), serviceFinalizer) &&
				slices.Contains(ue.ObjectNew.GetFinalizers(), serviceFinalizer))
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		WithEventFilter(filterPredicate).
		Complete(r)
}
