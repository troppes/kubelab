/*
Copyright 2023.

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

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubelabv1 "kubelab.local/kubelab/api/v1"
)

// KubelabUserReconciler reconciles a KubelabUser object
type KubelabUserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Definitions to manage status conditions
const (
	// typeAvailableMemcached represents the status of the Deployment reconciliation
	typeAvailableUser = "Available"
	// typeDegradedMemcached represents the status used when the custom resource is deleted and the finalizer operations are must to occur.
	typeDegradedUser = "Degraded"
)

const userFinalizer = "kubeuser.kubelab.local/finalizer"

//+kubebuilder:rbac:groups=kubelab.kubelab.local,resources=kubelabusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubelab.kubelab.local,resources=kubelabusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubelab.kubelab.local,resources=kubelabusers/finalizers,verbs=update

// RBAC group to create and delete namespaces
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *KubelabUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	// Fetch the Memcached instance
	// The purpose is check if the Custom Resource for the Kind Memcached
	// is applied on the cluster if not we return nil to stop the reconciliation
	user := &kubelabv1.KubelabUser{}
	err := r.Get(ctx, req.NamespacedName, user)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then, it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.Info("memcached resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get memcached")
		return ctrl.Result{}, err
	}

	// Let's just set the status as Unknown when no status are available
	if user.Status.Conditions == nil || len(user.Status.Conditions) == 0 {
		meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeAvailableUser, Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})
		if err = r.Status().Update(ctx, user); err != nil {
			log.Error(err, "Failed to update Memcached status")
			return ctrl.Result{}, err
		}

		// Let's re-fetch the memcached Custom Resource after update the status
		// so that we have the latest state of the resource on the cluster and we will avoid
		// raise the issue "the object has been modified, please apply
		// your changes to the latest version and try again" which would re-trigger the reconciliation
		// if we try to update it again in the following operations
		if err := r.Get(ctx, req.NamespacedName, user); err != nil {
			log.Error(err, "Failed to re-fetch memcached")
			return ctrl.Result{}, err
		}
	}

	// Let's add a finalizer. Then, we can define some operations which should
	// occurs before the custom resource to be deleted.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers
	if !controllerutil.ContainsFinalizer(user, userFinalizer) {
		log.Info("Adding Finalizer for Memcached")
		if ok := controllerutil.AddFinalizer(user, userFinalizer); !ok {
			log.Error(err, "Failed to add finalizer into the custom resource")
			return ctrl.Result{Requeue: true}, nil
		}

		if err = r.Update(ctx, user); err != nil {
			log.Error(err, "Failed to update custom resource to add finalizer")
			return ctrl.Result{}, err
		}
	}

	// Check if the Memcached instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isUserMarkedToBeDeleted := user.GetDeletionTimestamp() != nil
	if isUserMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(user, userFinalizer) {
			log.Info("Performing Finalizer Operations for User before delete CR")

			// Let's add here an status "Degraded" to define that this resource begin its process to be terminated.
			meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeDegradedUser,
				Status: metav1.ConditionUnknown, Reason: "Finalizing",
				Message: fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", user.Name)})

			if err := r.Status().Update(ctx, user); err != nil {
				log.Error(err, "Failed to update user status")
				return ctrl.Result{}, err
			}

			// Perform all operations required before remove the finalizer and allow
			// the Kubernetes API to remove the custom resource.
			r.doFinalizerOperations(user)

			// TODO(user): If you add operations to the doFinalizerOperationsForMemcached method
			// then you need to ensure that all worked fine before deleting and updating the Downgrade status
			// otherwise, you should requeue here.

			// Re-fetch the memcached Custom Resource before update the status
			// so that we have the latest state of the resource on the cluster and we will avoid
			// raise the issue "the object has been modified, please apply
			// your changes to the latest version and try again" which would re-trigger the reconciliation
			if err := r.Get(ctx, req.NamespacedName, user); err != nil {
				log.Error(err, "Failed to re-fetch user")
				return ctrl.Result{}, err
			}

			meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeDegradedUser,
				Status: metav1.ConditionTrue, Reason: "Finalizing",
				Message: fmt.Sprintf("Finalizer operations for custom resource %s name were successfully accomplished", user.Name)})

			if err := r.Status().Update(ctx, user); err != nil {
				log.Error(err, "Failed to update user status")
				return ctrl.Result{}, err
			}

			log.Info("Removing Finalizer for user after successfully perform the operations")
			if ok := controllerutil.RemoveFinalizer(user, userFinalizer); !ok {
				log.Error(err, "Failed to remove finalizer for user")
				return ctrl.Result{Requeue: true}, nil
			}

			if err := r.Update(ctx, user); err != nil {
				log.Error(err, "Failed to remove finalizer for user")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Check if the deployment already exists, if not create a new one
	ns := &v1.Namespace{}
	err = r.Get(ctx, client.ObjectKey{Name: user.Spec.Id}, ns)
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new deployment
		ns, err := r.namespaceForUser(user)
		if err != nil {
			log.Error(err, "Failed to define new NS resource for user")

			// The following implementation will update the status
			meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeAvailableUser,
				Status: metav1.ConditionFalse, Reason: "Reconciling",
				Message: fmt.Sprintf("Failed to create Deployment for the custom resource (%s): (%s)", user.Spec.Id, err)})

			if err := r.Status().Update(ctx, user); err != nil {
				log.Error(err, "Failed to update user status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		log.Info("Creating a new NS", "Namespace Name", ns.Name)
		if err = r.Create(ctx, ns); err != nil {
			log.Error(err, "Failed to create new Namespace", "Namespace Name", ns.Name)
			return ctrl.Result{}, err
		}

		// Namespace created successfully
		// We will requeue the reconciliation so that we can ensure the state
		// and move forward for the next operations
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Namespace")
		// Let's return the error for the reconciliation be re-trigged again
		return ctrl.Result{}, err
	}

	// The following implementation will update the status
	meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeAvailableUser,
		Status: metav1.ConditionTrue, Reason: "Reconciling",
		Message: fmt.Sprintf("Namespace for custom resource (%s) created successfully", user.Spec.Id)})

	if err := r.Status().Update(ctx, user); err != nil {
		log.Error(err, "Failed to update user status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// finalizeMemcached will perform the required operations before delete the CR.
func (r *KubelabUserReconciler) doFinalizerOperations(cr *kubelabv1.KubelabUser) {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.

	// Note: It is not recommended to use finalizers with the purpose of delete resources which are
	// created and managed in the reconciliation. These ones, such as the Deployment created on this reconcile,
	// are defined as depended of the custom resource. See that we use the method ctrl.SetControllerReference.
	// to set the ownerRef which means that the Deployment will be deleted by the Kubernetes API.
	// More info: https://kubernetes.io/docs/tasks/administer-cluster/use-cascading-deletion/
}

// deploymentForMemcached returns a Memcached Deployment object
func (r *KubelabUserReconciler) namespaceForUser(user *kubelabv1.KubelabUser) (*v1.Namespace, error) {
	ls := labelsForUser(user.Spec.Id)

	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   user.Spec.Id,
			Labels: ls,
		},
	}

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(user, ns, r.Scheme); err != nil {
		return nil, err
	}
	return ns, nil
}

// labelsForMemcached returns the labels for selecting the resources
// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
func labelsForUser(name string) map[string]string {

	return map[string]string{
		"app.kubernetes.io/name":       "KubelabUser",
		"app.kubernetes.io/instance":   name,
		"app.kubernetes.io/version":    "1",
		"app.kubernetes.io/part-of":    "kubelabuser-operator",
		"app.kubernetes.io/created-by": "controller-manager",
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubelabUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubelabv1.KubelabUser{}).
		Owns(&v1.Namespace{}).
		Complete(r)
}
