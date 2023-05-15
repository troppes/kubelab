package controller

import (
	v1apps "k8s.io/api/apps/v1"
	kubelabv1 "kubelab.local/kubelab/api/v1"
)

func isInClass(students []kubelabv1.KubelabUser, deployment v1apps.Deployment) bool {
	for _, student := range students {
		if student.Spec.Id == deployment.Namespace {
			return true
		}
	}
	return false
}
