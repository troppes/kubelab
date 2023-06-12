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
	"errors"
	"fmt"
	"strings"
	"time"

	v1apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
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

// ClassroomReconciler reconciles a Classroom object
type ClassroomReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubelab.kubelab.local,resources=classrooms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubelab.kubelab.local,resources=classrooms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubelab.kubelab.local,resources=classrooms/finalizers,verbs=update
//+kubebuilder:rbac:groups=kubelab.kubelab.local,resources=kubelabusers,verbs=get;list;watch

//Custom RBAC
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete

func (r *ClassroomReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the instance and check if it exist
	classroom := &kubelabv1.Classroom{}
	if err := r.Get(ctx, req.NamespacedName, classroom); err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then, it usually means that it was deleted or not created
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Class")
		return ctrl.Result{}, err
	}

	// set the status as Unknown when no status are available
	if classroom.Status.Conditions == nil || len(classroom.Status.Conditions) == 0 {
		meta.SetStatusCondition(&classroom.Status.Conditions, metav1.Condition{Type: typeAvailable, Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})
		if err := r.Status().Update(ctx, classroom); err != nil {
			log.Error(err, "Failed to update classroom status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Finalizer to ensure deletion of NS
	if classroom.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(classroom, classroomFinalizer) {
			controllerutil.AddFinalizer(classroom, classroomFinalizer)
			if err := r.Update(ctx, classroom); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(classroom, classroomFinalizer) {

			// Let's add here an status "Degraded" to define that this resource begin its process to be terminated.
			meta.SetStatusCondition(&classroom.Status.Conditions, metav1.Condition{Type: typeDegraded,
				Status: metav1.ConditionUnknown, Reason: "Finalizing",
				Message: fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", classroom.Name)})

			if err := r.Status().Update(ctx, classroom); err != nil {
				log.Error(err, "Failed to update classroom status")
				return ctrl.Result{}, err
			}

			// Since the Owners Reference does not delete the Namespace the Finalizer is used
			namespaceName := classroom.Name
			ns := &v1.Namespace{}
			err := r.Get(ctx, client.ObjectKey{Name: namespaceName}, ns)
			if err == nil {
				err = r.Delete(ctx, ns)
				if err != nil {
					log.Error(err, "Failed to delete Namespace", "Namespace Name", namespaceName)
					return ctrl.Result{}, err
				}
			}

			meta.SetStatusCondition(&classroom.Status.Conditions, metav1.Condition{Type: typeDegraded,
				Status: metav1.ConditionTrue, Reason: "Finalizing",
				Message: fmt.Sprintf("Finalizer operations for custom resource %s name were successfully accomplished", classroom.Name)})

			if err := r.Status().Update(ctx, classroom); err != nil {
				log.Error(err, "Failed to update classroom status")
				return ctrl.Result{}, err
			}

			log.Info("Successfully finalized classroom")
			if ok := controllerutil.RemoveFinalizer(classroom, classroomFinalizer); !ok {
				log.Error(err, "Failed to remove finalizer for classroom")
				return ctrl.Result{}, nil
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(classroom, classroomFinalizer)
			if err := r.Update(ctx, classroom); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// Check validity of connected ressources TO BE REPLACED WITH A VALIDATION WEBHOOK
	teacher := classroom.Spec.Teacher
	students := classroom.Spec.EnrolledStudents
	if teacher.Spec.Id == "" {
		return ctrl.Result{RequeueAfter: time.Minute}, errors.New("teacher not set")
	} else {
		kubelabUserList := &kubelabv1.KubelabUserList{}
		r.Client.List(ctx, kubelabUserList)

		for i := 0; i < len(students); i++ {
			if err := r.List(ctx, kubelabUserList, client.MatchingFields{userOwnerKey: students[i].Spec.Id}); err != nil || len(kubelabUserList.Items) == 0 {
				return ctrl.Result{RequeueAfter: time.Minute}, errors.New("student does not exist: " + students[i].Spec.Id)
			}
		}

		if err := r.List(ctx, kubelabUserList, client.MatchingFields{userOwnerKey: teacher.Spec.Id}); err != nil || len(kubelabUserList.Items) == 0 {
			return ctrl.Result{RequeueAfter: time.Minute}, errors.New("teacher does not exist: " + teacher.Spec.Id)
		} else if !kubelabUserList.Items[0].Spec.IsTeacher {
			return ctrl.Result{RequeueAfter: time.Minute}, errors.New("user is not a teacher: " + teacher.Spec.Id)
		}
	}

	// create a NS for class and Mount
	ns := &v1.Namespace{}
	if err := r.Get(ctx, client.ObjectKey{Name: classroom.Name}, ns); err != nil && apierrors.IsNotFound(err) {
		// Define a new NS
		ns, err := r.namespaceForClass(classroom)

		if err != nil {
			log.Error(err, "Failed to define new NS resource for classroom")

			// The following implementation will update the status
			meta.SetStatusCondition(&classroom.Status.Conditions, metav1.Condition{Type: typeAvailable,
				Status: metav1.ConditionFalse, Reason: "Reconciling",
				Message: fmt.Sprintf("Failed to create NS for the custom resource (%s): (%s)", classroom.Name, err)})

			if err := r.Status().Update(ctx, classroom); err != nil {
				log.Error(err, "Failed to update classroom status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		if err = r.Create(ctx, ns); err != nil {
			log.Error(err, "Failed to create new Namespace", "Namespace Name", ns.Namespace)
			return ctrl.Result{}, err
		}

		// requeue the reconciliation so that we can ensure the state
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Namespace")
		// Return the error for the reconciliation be re-trigged again
		return ctrl.Result{}, err
	}

	// Do operations for all students
	for _, student := range students {
		// Check if the deployment already exists, if not create a new one
		deployment := &v1apps.Deployment{}
		err := r.Get(ctx, types.NamespacedName{Name: classroom.Name, Namespace: student.Spec.Id}, deployment)
		if err != nil && apierrors.IsNotFound(err) {

			// fetch full student object
			studentList := &kubelabv1.KubelabUserList{}
			if err := r.List(ctx, studentList, client.MatchingFields{userOwnerKey: student.Spec.Id}); err != nil || len(studentList.Items) == 0 {
				return ctrl.Result{}, errors.New("unable to find Student")
			}

			// Define a new deployment
			dep, err := r.deploymentForClassroom(classroom, &studentList.Items[0])
			// If failing write Error inside Status
			if err != nil {
				log.Error(err, "Failed to define new Deployment resource for Classroom")

				// The following implementation will update the status
				meta.SetStatusCondition(&classroom.Status.Conditions, metav1.Condition{Type: typeAvailable,
					Status: metav1.ConditionFalse, Reason: "Reconciling",
					Message: fmt.Sprintf("Failed to create Deployment for the custom resource (%s): (%s)", classroom.Name, err)})

				if err := r.Status().Update(ctx, classroom); err != nil {
					log.Error(err, "Failed to update status")
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, err
			}

			if err = r.Create(ctx, dep); err != nil {
				log.Error(err, "Failed to create new Deployment",
					"Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
				return ctrl.Result{}, err
			}

			// Reque to check if everything is alright
			return ctrl.Result{RequeueAfter: time.Second * 10}, nil
		} else if err != nil {
			log.Error(err, "Failed to get Deployment")
			return ctrl.Result{}, err
		}

		// If the image gets changed in the CRD all deployments need to exchange theirs as well
		image := classroom.Spec.TemplateContainer
		// important: the call only works on the first image, so multiple images are currently not supported
		if deployment.Spec.Template.Spec.Containers[0].Image != image {
			deployment.Spec.Template.Spec.Containers[0].Image = image
			if err = r.Update(ctx, deployment); err != nil {
				log.Error(err, "Failed to update Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)

				// The following implementation will update the status
				meta.SetStatusCondition(&classroom.Status.Conditions, metav1.Condition{Type: typeDegraded,
					Status: metav1.ConditionFalse, Reason: "Changing Image",
					Message: fmt.Sprintf("Failed to update the image for the custom resource (%s): (%s)", classroom.Name, err)})

				if err := r.Status().Update(ctx, classroom); err != nil {
					log.Error(err, "Failed to update status")
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, err
			}

			// Now, that we update the image we want to requeue the reconciliation
			return ctrl.Result{Requeue: true}, nil
		}

		// Check if the svc already exists, if not create a new one
		service := &v1.Service{}
		err = r.Get(ctx, types.NamespacedName{Name: classroom.Name, Namespace: student.Spec.Id}, service)
		if err != nil && apierrors.IsNotFound(err) {
			// Define a new svc
			svc, err := r.serviceForClassroom(classroom, &student)
			// If failing write Error inside Status
			if err != nil {
				log.Error(err, "Failed to define new SVC resource for Classroom")

				// The following implementation will update the status
				meta.SetStatusCondition(&classroom.Status.Conditions, metav1.Condition{Type: typeAvailable,
					Status: metav1.ConditionFalse, Reason: "Reconciling",
					Message: fmt.Sprintf("Failed to create SVC for the custom resource (%s): (%s)", classroom.Name, err)})

				if err := r.Status().Update(ctx, classroom); err != nil {
					log.Error(err, "Failed to update status")
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, err
			}

			if err = r.Create(ctx, svc); err != nil {
				log.Error(err, "Failed to create new Deployment",
					"Deployment.Namespace", svc.Namespace, "SVC.Name", svc.Name)
				return ctrl.Result{}, err
			}

			// Reque to check if everything is alright
			return ctrl.Result{RequeueAfter: time.Second * 10}, nil
		} else if err != nil {
			log.Error(err, "Failed to get SVC")
			return ctrl.Result{}, err
		}

		np := &networkingv1.NetworkPolicy{}
		isExam := strings.ToLower(classroom.Spec.EnableExamMode) == "true"
		err = r.Get(ctx, types.NamespacedName{Name: classroom.Name, Namespace: student.Spec.Id}, np)
		if err != nil && apierrors.IsNotFound(err) && isExam {
			// Define a new network policy
			np, err := r.networkPolicyForClassroom(classroom, &student)
			// If failing write Error inside Status
			if err != nil {
				log.Error(err, "Failed to define new NP resource for Classroom")

				// The following implementation will update the status
				meta.SetStatusCondition(&classroom.Status.Conditions, metav1.Condition{Type: typeAvailable,
					Status: metav1.ConditionFalse, Reason: "Reconciling",
					Message: fmt.Sprintf("Failed to create NP for the custom resource (%s): (%s)", classroom.Name, err)})

				if err := r.Status().Update(ctx, classroom); err != nil {
					log.Error(err, "Failed to update status")
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, err
			}

			if err = r.Create(ctx, np); err != nil {
				log.Error(err, "Failed to create new NP")
				return ctrl.Result{}, err
			}

			// Reque to check if everything is alright
			return ctrl.Result{RequeueAfter: time.Second * 10}, nil
		} else if !isExam && err == nil { // check if it is not exam and it is found -> delete
			if err := r.Delete(ctx, np); err != nil {
				log.Error(err, "unable to delete network policy")
				return ctrl.Result{}, err
			} else {
				log.Info("Deleted NP", "Namespace", np.Namespace)
			}
			return ctrl.Result{}, err
		} else if err != nil && !apierrors.IsNotFound(err) { // it another error than not found
			log.Error(err, "Failed to get NetworkPolicy")
			return ctrl.Result{}, err
		}

	}

	// delete if student is removed
	deploymentList := &v1apps.DeploymentList{}
	if err := r.List(ctx, deploymentList, client.MatchingFields{classroomOwnerKey: classroom.Name}); err != nil {
		log.Error(err, "unable to list all child deployments")
		return ctrl.Result{}, err
	} else {
		for _, deploy := range deploymentList.Items {
			if !isInClass(students, deploy) {
				if err := r.Delete(ctx, &deploy); err != nil {
					log.Error(err, "unable to delete old deployment")
					return ctrl.Result{}, err
				} else {
					return ctrl.Result{}, err
				}
			}
		}
	}

	// Check if the Claim already exists, if not create a new Claim
	claim := &v1.PersistentVolumeClaim{}
	if err := r.Get(ctx, client.ObjectKey{Name: claimNameClass, Namespace: classroom.Name}, claim); err != nil && apierrors.IsNotFound(err) {
		// Define a new Role
		claim, err := r.persistentVolumeClaimForClassroom(classroom)

		if err != nil {
			log.Error(err, "Failed to define new PVC resource for classroom")

			// The following implementation will update the status
			meta.SetStatusCondition(&classroom.Status.Conditions, metav1.Condition{Type: typeAvailable,
				Status: metav1.ConditionFalse, Reason: "Reconciling",
				Message: fmt.Sprintf("Failed to create PVC for the custom resource (%s): (%s)", classroom.Name, err)})

			if err := r.Status().Update(ctx, classroom); err != nil {
				log.Error(err, "Failed to update status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		if err = r.Create(ctx, claim); err != nil {
			log.Error(err, "Failed to create new PVC")
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if err != nil {
		log.Error(err, "Failed to get PVC")
		// Return the error for the reconciliation be re-trigged again
		return ctrl.Result{}, err
	}

	// The following implementation will update the status
	meta.SetStatusCondition(&classroom.Status.Conditions, metav1.Condition{Type: typeAvailable,
		Status: metav1.ConditionTrue, Reason: "Reconciling",
		Message: fmt.Sprintf("Everything for custom resource (%s) created successfully", classroom.Name)})

	if err := r.Status().Update(ctx, classroom); err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClassroomReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1apps.Deployment{}, classroomOwnerKey, func(rawObj client.Object) []string {
		deploy := rawObj.(*v1apps.Deployment)
		return []string{deploy.Name}
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &kubelabv1.KubelabUser{}, userOwnerKey, func(rawObj client.Object) []string {
		user := rawObj.(*kubelabv1.KubelabUser)
		return []string{user.Spec.Id}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&kubelabv1.Classroom{}).
		Owns(&kubelabv1.KubelabUser{}).
		Owns(&v1apps.Deployment{}).
		Owns(&v1.Namespace{}).
		Owns(&v1.Service{}).
		Owns(&v1.PersistentVolumeClaim{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Complete(r)
}
