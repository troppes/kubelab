package controller

import (
	v1 "k8s.io/api/core/v1"
	v1rbac "k8s.io/api/rbac/v1"
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
				Resources: []string{"pods"},
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