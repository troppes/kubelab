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

// ClassroomSpec defines the desired state of Classroom
type ClassroomSpec struct {
	Teacher           KubelabUser   `json:"teacher,omitempty"`
	Namespace         string        `json:"namespace,omitempty"`
	EnrolledStudents  []KubelabUser `json:"enrolledStudents,omitempty"`
	TemplateContainer string        `json:"templateContainer,omitempty"`
}

// ClassroomStatus defines the observed state of Classroom
type ClassroomStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// Classroom is the Schema for the classrooms API
type Classroom struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClassroomSpec   `json:"spec,omitempty"`
	Status ClassroomStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClassroomList contains a list of Classroom
type ClassroomList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Classroom `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Classroom{}, &ClassroomList{})
}
