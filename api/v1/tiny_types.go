/*
Copyright 2021.

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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TinySpec defines the desired state of Tiny
type TinySpec struct {
	GPU       bool      `json:"gpu"`
	Framework Framework `json:"framework"`
}

// TinyStatus defines the observed state of Tiny
type TinyStatus struct {
	Phase    v1.PodPhase `json:"phase,omitempty"`
	NodePort int32       `json:"nodePort,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Tiny is the Schema for the tinies API
type Tiny struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TinySpec   `json:"spec,omitempty"`
	Status TinyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TinyList contains a list of Tiny
type TinyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tiny `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tiny{}, &TinyList{})
}
