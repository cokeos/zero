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

const (
	ResourceNvidiaGPU v1.ResourceName = "nvidia.com/gpu"
)

// UnitSpec defines the desired state of Unit
type UnitSpec struct {
	GPUPolicy    GPUPolicy          `json:"gpuPolicy"`
	Framework    Framework          `json:"framework"`
	ResourceList v1.ResourceList    `json:"resourceList"`
	Ports        []v1.ContainerPort `json:"ports,omitempty"`
	Execution    Execution          `json:"execution"`
}

type GPUPolicy struct {
	GPU    bool   `json:"gpu"`
	Model  string `json:"model,omitempty"`
	Number int    `json:"number"`
}

type Execution struct {
	SSH     bool        `json:"ssh"`
	Env     []v1.EnvVar `json:"env,omitempty"`
	Command []string    `json:"command,omitempty"`
	Args    []string    `json:"args,omitempty"`
}

type Framework struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// UnitStatus defines the observed state of Unit
type UnitStatus struct {
	Phase v1.PodPhase `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Unit is the Schema for the units API
type Unit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UnitSpec   `json:"spec,omitempty"`
	Status UnitStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// UnitList contains a list of Unit
type UnitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Unit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Unit{}, &UnitList{})
}
