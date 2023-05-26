package controller

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubelabv1 "kubelab.local/kubelab/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	v1apps "k8s.io/api/apps/v1"
)

// namespaceForClass returns a namespace for the Kubelabuser.
func (r *ClassroomReconciler) namespaceForClass(classroom *kubelabv1.Classroom) (*v1.Namespace, error) {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: classroom.Spec.Namespace,
		},
	}

	// Set the ownerRef for the Namespace for deletion of dependent, which does not seem to work with NS
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(classroom, ns, r.Scheme); err != nil {
		return nil, err
	}
	return ns, nil
}

// deploymentForClassroom returns a service object.
func (r *ClassroomReconciler) serviceForClassroom(classroom *kubelabv1.Classroom, student *kubelabv1.KubelabUser) (*v1.Service, error) {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      classroom.Spec.Namespace,
			Namespace: student.Spec.Id,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Ports: []v1.ServicePort{
				{
					Port: 2222,
					// TargetPort: intstr.FromInt(2222), defaults to port if not set
					// NodePort:   30000, // Randomly assigned if not set
				},
			},
			Selector: map[string]string{
				"class":   classroom.Spec.Namespace,
				"student": student.Spec.Id,
			},
		},
	}

	// Set the ownerRef
	if err := ctrl.SetControllerReference(classroom, service, r.Scheme); err != nil {
		return nil, err
	}
	return service, nil
}

// deploymentForClassroom returns a Deployment object.
func (r *ClassroomReconciler) deploymentForClassroom(classroom *kubelabv1.Classroom, student *kubelabv1.KubelabUser) (*v1apps.Deployment, error) {
	ls := labelsForClassroom(classroom.Spec.Namespace, student.Spec.Id)
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
					Containers: []v1.Container{{
						Image:           classroom.Spec.TemplateContainer,
						Name:            classroom.Spec.Namespace,
						ImagePullPolicy: v1.PullIfNotPresent,
						Ports: []v1.ContainerPort{{
							ContainerPort: 2222,
							Name:          "classroom-port",
						}},
						Env: []v1.EnvVar{
							{
								Name:  "PASSWORD_ACCESS",
								Value: "true",
							},
							{
								Name:  "SUDO_ACCESS",
								Value: "true",
							},
							{
								Name:  "USER_NAME",
								Value: "admin",
							},
							{
								Name:  "USER_PASSWORD",
								Value: "admin",
							},
						},
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
