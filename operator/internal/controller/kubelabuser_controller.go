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
	v1rbac "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

//+kubebuilder:rbac:groups=kubelab.kubelab.local,resources=kubelabusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubelab.kubelab.local,resources=kubelabusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubelab.kubelab.local,resources=kubelabusers/finalizers,verbs=update

// RBAC group to create and delete namespaces
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

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
		return ctrl.Result{}, nil
	}

	// Finalizer to ensure deletion of NS
	if user.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(user, userFinalizer) {
			controllerutil.AddFinalizer(user, userFinalizer)
			if err := r.Update(ctx, user); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(user, userFinalizer) {

			// Let's add here an status "Degraded" to define that this resource begin its process to be terminated.
			meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeDegraded,
				Status: metav1.ConditionUnknown, Reason: "Finalizing",
				Message: fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", user.Name)})

			if err := r.Status().Update(ctx, user); err != nil {
				log.Error(err, "Failed to update user status")
				return ctrl.Result{}, err
			}

			// Since the Owners Reference does not delete the Namespace the Finalizer is used
			ns := &v1.Namespace{}
			if err := r.Get(ctx, client.ObjectKey{Name: user.Spec.Id}, ns); err == nil {
				err = r.Delete(ctx, ns)
				if err != nil {
					log.Error(err, "Failed to delete Namespace", "Name", user.Spec.Id)
					return ctrl.Result{}, err
				}
			}
			if user.Spec.IsTeacher {
				roleBinding := &v1rbac.ClusterRoleBinding{}
				err := r.Get(ctx, client.ObjectKey{Name: kubelabPrefix + "teacher"}, roleBinding)
				if err == nil {
					err = r.Delete(ctx, roleBinding)
					if err != nil {
						log.Error(err, "Failed to delete Rolebinding", "Name", user.Spec.Id)
						return ctrl.Result{}, err
					}
				}

				role := &v1rbac.ClusterRole{}
				if err := r.Get(ctx, client.ObjectKey{Name: kubelabPrefix + "teacher"}, role); err == nil {
					err = r.Delete(ctx, role)
					if err != nil {
						log.Error(err, "Failed to delete Role", "Name", user.Spec.Id)
						return ctrl.Result{}, err
					}
				}
			}

			meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeDegraded,
				Status: metav1.ConditionTrue, Reason: "Finalizing",
				Message: fmt.Sprintf("Finalizer operations for custom resource %s name were successfully accomplished", user.Name)})

			if err := r.Status().Update(ctx, user); err != nil {
				log.Error(err, "Failed to update user status")
				return ctrl.Result{}, err
			}

			log.Info("Successfully finalized User")
			if ok := controllerutil.RemoveFinalizer(user, userFinalizer); !ok {
				log.Error(err, "Failed to remove finalizer for user")
				return ctrl.Result{Requeue: true}, nil
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(user, userFinalizer)
			if err := r.Update(ctx, user); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// Check if the NS already exists, if not create a new one and assign role
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
				Message: fmt.Sprintf("Failed to create NS for the custom resource (%s): (%s)", user.Spec.Id, err)})

			if err := r.Status().Update(ctx, user); err != nil {
				log.Error(err, "Failed to update user status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		if err = r.Create(ctx, ns); err != nil {
			log.Error(err, "Failed to create new Namespace", "Namespace Name", ns.Name)
			return ctrl.Result{}, err
		}

		// Namespace created successfully
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Namespace")
		// Return the error for the reconciliation be re-trigged again
		return ctrl.Result{}, err
	}

	// Check if the Role already exists, if not create a new one and add rolebinding
	role := &v1rbac.Role{}
	err = r.Get(ctx, types.NamespacedName{Name: roleName, Namespace: user.Spec.Id}, role)
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new Role
		role, err := r.roleForUser(user)

		if err != nil {
			log.Error(err, "Failed to define new Role resource for user")

			// The following implementation will update the status
			meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeAvailable,
				Status: metav1.ConditionFalse, Reason: "Reconciling",
				Message: fmt.Sprintf("Failed to create Role for the custom resource (%s): (%s)", user.Spec.Id, err)})

			if err := r.Status().Update(ctx, user); err != nil {
				log.Error(err, "Failed to update user status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		if err = r.Create(ctx, role); err != nil {
			log.Error(err, "Failed to create new Role")
			return ctrl.Result{}, err
		}
		// Requeue after 2 seconds to create Rolebinding
		return ctrl.Result{RequeueAfter: time.Second * 2}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Role")
		// Return the error for the reconciliation be re-trigged again
		return ctrl.Result{}, err
	}

	// Check if the Role already exists, if not create a new one and add rolebinding
	roleBinding := &v1rbac.RoleBinding{}
	err = r.Get(ctx, types.NamespacedName{Name: roleBindingName, Namespace: user.Spec.Id}, roleBinding)
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new Role
		roleBinding, err := r.rolebindingForUser(user)

		if err != nil {
			log.Error(err, "Failed to define new Rolebinding resource for user")

			// The following implementation will update the status
			meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeAvailable,
				Status: metav1.ConditionFalse, Reason: "Reconciling",
				Message: fmt.Sprintf("Failed to create Rolebinding for the custom resource (%s): (%s)", user.Spec.Id, err)})

			if err := r.Status().Update(ctx, user); err != nil {
				log.Error(err, "Failed to update user status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		if err = r.Create(ctx, roleBinding); err != nil {
			log.Error(err, "Failed to create new Rolebinding")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Rolebinding")
		// Return the error for the reconciliation be re-trigged again
		return ctrl.Result{}, err
	}

	// if User is a teacher give them the rights to list classes and students
	if user.Spec.IsTeacher {
		clusteRole := &v1rbac.ClusterRole{}
		if err := r.Get(ctx, client.ObjectKey{Name: kubelabPrefix + "teacher"}, clusteRole); err != nil && apierrors.IsNotFound(err) {
			// Define a new Role
			clusteRole, err := r.roleForTeacher(user)

			if err != nil {
				log.Error(err, "Failed to define new Role resource for user")

				// The following implementation will update the status
				meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeAvailable,
					Status: metav1.ConditionFalse, Reason: "Reconciling",
					Message: fmt.Sprintf("Failed to create Role for the custom resource (%s): (%s)", user.Name, err)})

				if err := r.Status().Update(ctx, user); err != nil {
					log.Error(err, "Failed to update user status")
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, err
			}

			if err = r.Create(ctx, clusteRole); err != nil {
				log.Error(err, "Failed to create new Role")
				return ctrl.Result{}, err
			}
			// Requeue after 2 seconds to create Rolebinding
			return ctrl.Result{RequeueAfter: time.Second * 2}, nil
		} else if err != nil {
			log.Error(err, "Failed to get Role")
			// Return the error for the reconciliation be re-trigged again
			return ctrl.Result{}, err
		}

		// Check if the Role already exists, if not create a new one and add rolebinding
		clusteRoleBinding := &v1rbac.ClusterRoleBinding{}
		if err := r.Get(ctx, client.ObjectKey{Name: kubelabPrefix + "teacher"}, clusteRoleBinding); err != nil && apierrors.IsNotFound(err) {
			// Define a new Role
			log.Info("REACHED")
			clusteRoleBinding, err := r.roleBindingForTeacher(user)

			if err != nil {
				log.Error(err, "Failed to define new Rolebinding resource for user")

				// The following implementation will update the status
				meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeAvailable,
					Status: metav1.ConditionFalse, Reason: "Reconciling",
					Message: fmt.Sprintf("Failed to create Rolebinding for the custom resource (%s): (%s)", user.Name, err)})

				if err := r.Status().Update(ctx, user); err != nil {
					log.Error(err, "Failed to update user status")
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, err
			}

			if err = r.Create(ctx, clusteRoleBinding); err != nil {
				log.Error(err, "Failed to create new Rolebinding")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		} else if err != nil {
			log.Error(err, "Failed to get Rolebinding")
			// Return the error for the reconciliation be re-trigged again
			return ctrl.Result{}, err
		}
	}

	// Check if the Claim already exists, if not create a new Claim
	claim := &v1.PersistentVolumeClaim{}
	err = r.Get(ctx, types.NamespacedName{Name: claimNameUser, Namespace: user.Spec.Id}, claim)
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new Role
		claim, err := r.persistentVolumeClaimForUser(user)

		if err != nil {
			log.Error(err, "Failed to define new PVC resource for user")

			// The following implementation will update the status
			meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeAvailable,
				Status: metav1.ConditionFalse, Reason: "Reconciling",
				Message: fmt.Sprintf("Failed to create PVC for the custom resource (%s): (%s)", user.Spec.Id, err)})

			if err := r.Status().Update(ctx, user); err != nil {
				log.Error(err, "Failed to update user status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		if err = r.Create(ctx, claim); err != nil {
			log.Error(err, "Failed to create new PVC")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get PVC")
		// Return the error for the reconciliation be re-trigged again
		return ctrl.Result{}, err
	}

	// The following implementation will update the status
	meta.SetStatusCondition(&user.Status.Conditions, metav1.Condition{Type: typeAvailable,
		Status: metav1.ConditionTrue, Reason: "Reconciling",
		Message: "Finished Reconciling"})

	if err := r.Status().Update(ctx, user); err != nil {
		log.Error(err, "Failed to update user status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubelabUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubelabv1.KubelabUser{}).
		Owns(&v1.Namespace{}).
		Owns(&v1rbac.Role{}).
		Owns(&v1rbac.RoleBinding{}).
		Owns(&v1.PersistentVolumeClaim{}).
		Complete(r)
}
