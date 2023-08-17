package controller

import (
	v1 "k8s.io/api/core/v1"
	v1rbac "k8s.io/api/rbac/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubelabv1 "kubelab.local/kubelab/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

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

// roleForUser returns role to scale and get ressources inside the namespace.
func (r *KubelabUserReconciler) roleForUser(user *kubelabv1.KubelabUser) (*v1rbac.Role, error) {

	// Define the Role object
	role := &v1rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName, // static names can be used here, since the namespace is unique
			Namespace: user.Spec.Id,
		},
		Rules: []v1rbac.PolicyRule{
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments"},
				Verbs:     []string{"list", "scale"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"get"},
			},
		},
	}

	if err := ctrl.SetControllerReference(user, role, r.Scheme); err != nil {
		return nil, err
	}

	return role, nil
}

// rolebindingForUser returns rolebinding to scale all ressources inside the namespace.
func (r *KubelabUserReconciler) rolebindingForUser(user *kubelabv1.KubelabUser) (*v1rbac.RoleBinding, error) {

	rb := &v1rbac.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleBindingName,
			Namespace: user.Spec.Id,
		},
		Subjects: []v1rbac.Subject{
			{
				Kind:      "Group",
				Name:      groupPrefix + user.Spec.Id,
				Namespace: user.Spec.Id,
			},
		},
		RoleRef: v1rbac.RoleRef{
			Kind:     "Role",
			Name:     roleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	if err := ctrl.SetControllerReference(user, rb, r.Scheme); err != nil {
		return nil, err
	}

	return rb, nil
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

// persistentVolumeClaimForUser returns pvc to have private folder.
func (r *KubelabUserReconciler) persistentVolumeClaimForUser(user *kubelabv1.KubelabUser) (*v1.PersistentVolumeClaim, error) {
	storageClassName := storageClass

	claim := &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      claimNameUser,
			Namespace: user.Spec.Id,
			Annotations: map[string]string{
				"nfs.io/storage-path": "student",
			},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClassName,
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteMany,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("100Mi"),
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(user, claim, r.Scheme); err != nil {
		return nil, err
	}

	return claim, nil
}

// roleForUser returns role to scale and get ressources inside the namespace.
func (r *KubelabUserReconciler) roleForTeacher(teacher *kubelabv1.KubelabUser) (*v1rbac.ClusterRole, error) {

	// Define the Role object
	role := &v1rbac.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubelabPrefix + "teacher",
		},
		Rules: []v1rbac.PolicyRule{
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments"},
				Verbs:     []string{"list", "scale", "update", "get"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{"kubelab.kubelab.local"},
				Resources: []string{"classrooms"},
				Verbs:     []string{"list"},
			},
		},
	}

	if err := ctrl.SetControllerReference(teacher, role, r.Scheme); err != nil {
		return nil, err
	}

	return role, nil
}

// roleBindingForTeacher returns rolebinding to give teachers the needed rights.
func (r *KubelabUserReconciler) roleBindingForTeacher(teacher *kubelabv1.KubelabUser) (*v1rbac.ClusterRoleBinding, error) {

	rb := &v1rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubelabPrefix + "teacher",
		},
		Subjects: []v1rbac.Subject{
			{
				Kind: "Group",
				Name: groupPrefix + "teacher",
			},
		},
		RoleRef: v1rbac.RoleRef{
			Kind:     "ClusterRole",
			Name:     kubelabPrefix + "teacher",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	if err := ctrl.SetControllerReference(teacher, rb, r.Scheme); err != nil {
		return nil, err
	}

	return rb, nil
}
