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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubelabUserSpec defines the desired state of KubelabUser
type KubelabUserSpec struct {
	// Normally StudentID, otherwise TeacherID
	Id        string `json:"id,omitempty"`
	IsTeacher bool   `json:"isTeacher,omitempty"`
}

// KubelabUserStatus defines the observed state of KubelabUser
type KubelabUserStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
	MountName  string             `json:"mountName,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// KubelabUser is the Schema for the kubelabusers API
type KubelabUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubelabUserSpec   `json:"spec,omitempty"`
	Status KubelabUserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubelabUserList contains a list of KubelabUser
type KubelabUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubelabUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubelabUser{}, &KubelabUserList{})
}
