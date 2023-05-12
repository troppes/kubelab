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

	v1apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

//Custom RBAC for Namespace and Deploy
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubelab.kubelab.local,resources=kubelabusers,verbs=get;list;watch

func (r *ClassroomReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("Classroom Created!")

	// Fetch the instance and check if it exist
	user := &kubelabv1.Classroom{}
	if err := r.Get(ctx, req.NamespacedName, user); err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then, it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Class")
		return ctrl.Result{}, err
	}

	// Check validity of connected ressources
	// only do for uncertain status?
	teacher := user.Spec.Teacher
	students := user.Spec.EnrolledStudents
	if teacher.Spec.Id == "" {
		err := errors.New("teacher not set")
		log.Error(err, "Teacher is an empty string")
		return ctrl.Result{}, err
	} else {
		kubelabuserList := &kubelabv1.KubelabUserList{}
		teacherFound := false
		r.Client.List(ctx, kubelabuserList)

		// map the students to id for better deletion
		studentMap := make(map[string]int)
		for i := 0; i < len(students); i++ {
			studentMap[students[i].Spec.Id] = 1
		}

		for _, user := range kubelabuserList.Items {
			if user.Spec.Id == teacher.Spec.Id {
				teacherFound = true
			}
			delete(studentMap, user.Spec.Id)
		}

		if !teacherFound {
			log.Error(nil, "Teacher does not exist")
			// TODO set requeue delay later
			return ctrl.Result{}, errors.New("teacher does not exist")
		}
		if len(studentMap) > 0 {
			log.Error(nil, "Not all students were found", "Students", studentMap)
			return ctrl.Result{}, errors.New("students do not exist")
		}
	}

	// create a NS for class and Mount

	// deploy in students NS

	// create Mount

	// on deletetion cascade and delete NS per Finalizer

	// Create Mount for Class and Mount (later)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClassroomReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubelabv1.Classroom{}).
		Owns(&kubelabv1.KubelabUser{}).
		Owns(&v1apps.Deployment{}).
		Owns(&v1.Namespace{}).
		Complete(r)
}
