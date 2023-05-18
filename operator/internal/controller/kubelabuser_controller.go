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

const userFinalizer = "kubeuser.kubelab.local/finalizer"

// KubelabUserReconciler reconciles a KubelabUser object
type KubelabUserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubelab.kubelab.local,resources=kubelabusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubelab.kubelab.local,resources=kubelabusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubelab.kubelab.local,resources=kubelabusers/finalizers,verbs=update

// RBAC group to create and delete namespaces
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete

func (r *KubelabUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	// Fetch the instance
	user := &kubelabv1.KubelabUser{}
	err := r.Get(ctx, req.NamespacedName, user)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then, it usually means that it was deleted or not created
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get User")
		return ctrl.Result{}, err
	}

	// set the status as Unknown when no status are available
	if user.Status.Conditions == nil || len(user.Status.Conditions) == 0 {
		meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeAvailable, Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})
		if err = r.Status().Update(ctx, user); err != nil {
			log.Error(err, "Failed to update User status")
			return ctrl.Result{}, err
		}

		// re-fetch the Custom Resource after update the status to ensure latest state
		if err := r.Get(ctx, req.NamespacedName, user); err != nil {
			log.Error(err, "Failed to re-fetch user")
			return ctrl.Result{}, err
		}
	}

	// Finalizer to ensure deletion of NS
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers
	if !controllerutil.ContainsFinalizer(user, userFinalizer) {
		log.Info("Adding Finalizer to User")
		if ok := controllerutil.AddFinalizer(user, userFinalizer); !ok {
			log.Error(err, "Failed to add finalizer into the custom resource")
			return ctrl.Result{Requeue: true}, nil
		}

		if err = r.Update(ctx, user); err != nil {
			log.Error(err, "Failed to update custom resource to add finalizer")
			return ctrl.Result{}, err
		}
	}

	// Check if the User instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isUserMarkedToBeDeleted := user.GetDeletionTimestamp() != nil
	if isUserMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(user, userFinalizer) {
			log.Info("Performing Finalizer Operations for User before delete CR")

			// Let's add here an status "Degraded" to define that this resource begin its process to be terminated.
			meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeDegraded,
				Status: metav1.ConditionUnknown, Reason: "Finalizing",
				Message: fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", user.Name)})

			if err := r.Status().Update(ctx, user); err != nil {
				log.Error(err, "Failed to update user status")
				return ctrl.Result{}, err
			}

			// Since the Owners Reference does not delete the Namespace the Finalizer is used
			namespaceName := user.Spec.Id
			ns := &v1.Namespace{}
			err := r.Get(ctx, client.ObjectKey{Name: namespaceName}, ns)
			if err == nil {
				err = r.Delete(ctx, ns)
				if err != nil {
					log.Error(err, "Failed to delete Namespace", "Namespace Name", namespaceName)
					return ctrl.Result{}, err
				}
				log.Info("Deleted Namespace", "Namespace Name", namespaceName)
			}

			// Re-fetch the Custom Resource after update the status to ensure latest state
			if err := r.Get(ctx, req.NamespacedName, user); err != nil {
				log.Error(err, "Failed to re-fetch user")
				return ctrl.Result{}, err
			}

			meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeDegraded,
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

	// Check if the NS already exists, if not create a new one
	ns := &v1.Namespace{}
	err = r.Get(ctx, client.ObjectKey{Name: user.Spec.Id}, ns)
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new NS
		ns, err := r.namespaceForUser(user)

		if err != nil {
			log.Error(err, "Failed to define new NS resource for user")

			// The following implementation will update the status
			meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeAvailable,
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
		// Return the error for the reconciliation be re-trigged again
		return ctrl.Result{}, err
	}

	// The following implementation will update the status
	meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeAvailable,
		Status: metav1.ConditionTrue, Reason: "Reconciling",
		Message: fmt.Sprintf("Namespace for custom resource (%s) created successfully", user.Spec.Id)})

	if err := r.Status().Update(ctx, user); err != nil {
		log.Error(err, "Failed to update user status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// namespaceForUser returns a namespace for the Kubelabuser
func (r *KubelabUserReconciler) namespaceForUser(user *kubelabv1.KubelabUser) (*v1.Namespace, error) {
	ls := labelsForUser(user.Spec.Id)

	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   user.Spec.Id,
			Labels: ls,
		},
	}

	// Set the ownerRef for the Namespace for deletion of dependent, which does not seem to work with NS
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(user, ns, r.Scheme); err != nil {
		return nil, err
	}
	return ns, nil
}

// labelsForUser returns the labels for selecting the resources
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
