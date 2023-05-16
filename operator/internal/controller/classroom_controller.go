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
	"time"

	v1apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
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

const classroomFinalizer = "classroom.kubelab.local/finalizer"
const classroomOwnerKey = ".metadata.namespace"
const userOwnerKey = ".spec.id"

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

	// Fetch the instance and check if it exist
	classroom := &kubelabv1.Classroom{}
	if err := r.Get(ctx, req.NamespacedName, classroom); err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then, it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Class")
		return ctrl.Result{}, err
	}

	// Let's just set the status as Unknown when no status are available
	if classroom.Status.Conditions == nil || len(classroom.Status.Conditions) == 0 {
		meta.SetStatusCondition(&classroom.Status.Conditions, metav1.Condition{Type: typeAvailable, Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})
		if err := r.Status().Update(ctx, classroom); err != nil {
			log.Error(err, "Failed to update classroom status")
			return ctrl.Result{}, err
		}

		// Let's re-fetch the Custom Resource after update the status to ensure latest state
		if err := r.Get(ctx, req.NamespacedName, classroom); err != nil {
			log.Error(err, "Failed to re-fetch classroom")
			return ctrl.Result{}, err
		}
	}

	// Finalizer to ensure deletion of NS
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers
	if !controllerutil.ContainsFinalizer(classroom, classroomFinalizer) {
		log.Info("Adding Finalizer to Classroom")
		if ok := controllerutil.AddFinalizer(classroom, classroomFinalizer); !ok {
			log.Error(errors.New("finalizer does not exist"), "Failed to add finalizer into the custom resource")
			return ctrl.Result{Requeue: true}, nil
		}

		if err := r.Update(ctx, classroom); err != nil {
			log.Error(err, "Failed to update custom resource to add finalizer")
			return ctrl.Result{}, err
		}
	}

	// Check if the User instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isClassroomMarkedToBeDeleted := classroom.GetDeletionTimestamp() != nil
	if isClassroomMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(classroom, classroomFinalizer) {
			log.Info("Performing Finalizer Operations for classroom before delete CR")

			// Let's add here an status "Degraded" to define that this resource begin its process to be terminated.
			meta.SetStatusCondition(&classroom.Status.Conditions, metav1.Condition{Type: typeDegraded,
				Status: metav1.ConditionUnknown, Reason: "Finalizing",
				Message: fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", classroom.Name)})

			if err := r.Status().Update(ctx, classroom); err != nil {
				log.Error(err, "Failed to update classroom status")
				return ctrl.Result{}, err
			}

			// Re-fetch the Custom Resource after update the status to ensure latest state
			if err := r.Get(ctx, req.NamespacedName, classroom); err != nil {
				log.Error(err, "Failed to re-fetch classroom")
				return ctrl.Result{}, err
			}

			// Since the Owners Reference does not delete the Namespace the Finalizer is used
			namespaceName := classroom.Spec.Namespace
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
			if err := r.Get(ctx, req.NamespacedName, classroom); err != nil {
				log.Error(err, "Failed to re-fetch classroom")
				return ctrl.Result{}, err
			}

			meta.SetStatusCondition(&classroom.Status.Conditions, metav1.Condition{Type: typeDegraded,
				Status: metav1.ConditionTrue, Reason: "Finalizing",
				Message: fmt.Sprintf("Finalizer operations for custom resource %s name were successfully accomplished", classroom.Name)})

			if err := r.Status().Update(ctx, classroom); err != nil {
				log.Error(err, "Failed to update classroom status")
				return ctrl.Result{}, err
			}

			log.Info("Removing Finalizer for classroom after successfully perform the operations")
			if ok := controllerutil.RemoveFinalizer(classroom, classroomFinalizer); !ok {
				log.Error(err, "Failed to remove finalizer for classroom")
				return ctrl.Result{Requeue: true}, nil
			}

			if err := r.Update(ctx, classroom); err != nil {
				log.Error(err, "Failed to remove finalizer for classroom")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Check validity of connected ressources
	teacher := classroom.Spec.Teacher
	students := classroom.Spec.EnrolledStudents
	if teacher.Spec.Id == "" {
		err := errors.New("teacher not set")
		log.Error(err, "Teacher is an empty string")
		return ctrl.Result{}, err
	} else {
		kubelabUserList := &kubelabv1.KubelabUserList{}
		r.Client.List(ctx, kubelabUserList)

		for i := 0; i < len(students); i++ {
			if err := r.List(ctx, kubelabUserList, client.MatchingFields{userOwnerKey: students[i].Spec.Id}); err != nil || len(kubelabUserList.Items) == 0 {
				return ctrl.Result{}, errors.New("student does not exist: " + students[i].Spec.Id)
			}
		}

		if err := r.List(ctx, kubelabUserList, client.MatchingFields{userOwnerKey: teacher.Spec.Id}); err != nil || len(kubelabUserList.Items) == 0 {
			return ctrl.Result{}, errors.New("teacher does not exist: " + teacher.Spec.Id)
		} else if !kubelabUserList.Items[0].Spec.IsTeacher {
			return ctrl.Result{}, errors.New("user is not a teacher: " + teacher.Spec.Id)
		}
	}

	// create a NS for class and Mount
	// Check if the NS already exists, if not create a new one
	ns := &v1.Namespace{}
	if err := r.Get(ctx, client.ObjectKey{Name: classroom.Spec.Namespace}, ns); err != nil && apierrors.IsNotFound(err) {
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

		log.Info("Creating a new NS", "Namespace Name", ns.Namespace)

		if err = r.Create(ctx, ns); err != nil {
			log.Error(err, "Failed to create new Namespace", "Namespace Name", ns.Namespace)
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

	// Create Deployments for all students, if not already existing
	for _, student := range students {
		// Check if the deployment already exists, if not create a new one
		found := &v1apps.Deployment{}
		err := r.Get(ctx, types.NamespacedName{Name: classroom.Name, Namespace: student.Spec.Id}, found)
		if err != nil && apierrors.IsNotFound(err) {
			// Define a new deployment
			dep, err := r.deploymentForClassroom(classroom, &student)
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

			log.Info("Creating a new Deployment",
				"Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			if err = r.Create(ctx, dep); err != nil {
				log.Error(err, "Failed to create new Deployment",
					"Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
				return ctrl.Result{}, err
			}

			// Deployment created successfully
			// Reque to check if everything is alright
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		} else if err != nil {
			log.Error(err, "Failed to get Deployment")
			// Let's return the error for the reconciliation be re-trigged again
			return ctrl.Result{}, err
		}

		// If the image gets changed in the CRD all deployments need to exchange theirs as well
		image := classroom.Spec.TemplateContainer
		// important: the call only works on the first image, so multiple images are currently not supported
		if found.Spec.Template.Spec.Containers[0].Image != image {
			found.Spec.Template.Spec.Containers[0].Image = image
			if err = r.Update(ctx, found); err != nil {
				log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)

				// Re-fetch to ensure state
				if err := r.Get(ctx, req.NamespacedName, classroom); err != nil {
					log.Error(err, "Failed to re-fetch")
					return ctrl.Result{}, err
				}

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
			// so that we can ensure that we have the latest state of the resource before
			// update. Also, it will help ensure the desired state on the cluster
			return ctrl.Result{Requeue: true}, nil
		}
	}

	// delete if student is removed
	deploymentList := &v1apps.DeploymentList{}
	if err := r.List(ctx, deploymentList, client.MatchingFields{classroomOwnerKey: classroom.Spec.Namespace}); err != nil {
		log.Error(err, "unable to list all child deployments")
		return ctrl.Result{}, err
	} else {
		for _, deploy := range deploymentList.Items {
			if !isInClass(students, deploy) {
				if err := r.Delete(ctx, &deploy); err != nil {
					log.Error(err, "unable to delete old deployment")
					return ctrl.Result{}, err
				} else {
					log.Info("Deleted Deployment", "Deployment Namespace", deploy.Namespace)
				}
			}
		}
	}

	// Create Mount for Class (later)

	// The following implementation will update the status
	meta.SetStatusCondition(&classroom.Status.Conditions, metav1.Condition{Type: typeAvailable,
		Status: metav1.ConditionTrue, Reason: "Reconciling",
		Message: fmt.Sprintf("Namespace and deployment for custom resource (%s) created successfully", classroom.Name)})

	if err := r.Status().Update(ctx, classroom); err != nil {
		log.Error(err, "Failed to update user status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// namespaceForClass returns a namespace for the Kubelabuser.
func (r *ClassroomReconciler) namespaceForClass(classroom *kubelabv1.Classroom) (*v1.Namespace, error) {
	ls := labelsForClassroom(classroom.Name)

	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   classroom.Spec.Namespace,
			Labels: ls,
		},
	}

	// Set the ownerRef for the Namespace for deletion of dependent, which does not seem to work with NS
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(classroom, ns, r.Scheme); err != nil {
		return nil, err
	}
	return ns, nil
}

// deploymentForClassroom returns a Deployment object.
func (r *ClassroomReconciler) deploymentForClassroom(classroom *kubelabv1.Classroom, student *kubelabv1.KubelabUser) (*v1apps.Deployment, error) {
	ls := labelsForClassroom(classroom.Name)
	replicas := int32(1)

	deployment := &v1apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      classroom.Spec.Namespace,
			Namespace: student.Spec.Id,
		},
		Spec: v1apps.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: v1.PodSpec{
					// let only run on linux for now
					Affinity: &v1.Affinity{
						NodeAffinity: &v1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
								NodeSelectorTerms: []v1.NodeSelectorTerm{
									{
										MatchExpressions: []v1.NodeSelectorRequirement{
											{
												Key:      "kubernetes.io/arch",
												Operator: "In",
												Values:   []string{"amd64", "arm64", "ppc64le", "s390x"},
											},
											{
												Key:      "kubernetes.io/os",
												Operator: "In",
												Values:   []string{"linux"},
											},
										},
									},
								},
							},
						},
					},
					SecurityContext: &v1.PodSecurityContext{
						RunAsNonRoot: &[]bool{true}[0],
						// TODO LEARN WHAT EXACTLY IT DOES
						SeccompProfile: &v1.SeccompProfile{
							Type: v1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Containers: []v1.Container{{
						Image:           classroom.Spec.TemplateContainer,
						Name:            classroom.Spec.Namespace,
						ImagePullPolicy: v1.PullIfNotPresent,
						// Ensure restrictive context for the container
						// More info: https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted
						SecurityContext: &v1.SecurityContext{
							// WARNING: Ensure that the image used defines an UserID in the Dockerfile
							// otherwise the Pod will not run and will fail with "container has runAsNonRoot and image has non-numeric user"".
							// If you want your workloads admitted in namespaces enforced with the restricted mode in OpenShift/OKD vendors
							// then, you MUST ensure that the Dockerfile defines a User ID OR you MUST leave the "RunAsNonRoot" and
							// "RunAsUser" fields empty.
							RunAsNonRoot: &[]bool{true}[0],
							// The memcached image does not use a non-zero numeric user as the default user.
							// Due to RunAsNonRoot field being set to true, we need to force the user in the
							// container to a non-zero numeric user. We do this using the RunAsUser field.
							// However, if you are looking to provide solution for K8s vendors like OpenShift
							// be aware that you cannot run under its restricted-v2 SCC if you set this value.
							RunAsUser:                &[]int64{1001}[0],
							AllowPrivilegeEscalation: &[]bool{false}[0],
							Capabilities: &v1.Capabilities{
								Drop: []v1.Capability{
									"ALL",
								},
							},
						},
						Ports: []v1.ContainerPort{{
							// TODO change Port
							ContainerPort: 4222,
							Name:          "classroom-port",
						}},
						Command: []string{"memcached", "-m=64", "-o", "modern", "-v"},
					}},
				},
			},
		},
	}

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(classroom, deployment, r.Scheme); err != nil {
		return nil, err
	}
	return deployment, nil
}

// labelsForClassroom returns the labels for selecting the resources.
func labelsForClassroom(name string) map[string]string {

	return map[string]string{
		"app.kubernetes.io/name":       "KubelabClassroom",
		"app.kubernetes.io/instance":   name,
		"app.kubernetes.io/version":    "1",
		"app.kubernetes.io/part-of":    "classroom-operator",
		"app.kubernetes.io/created-by": "controller-manager",
	}
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
		Complete(r)
}
