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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Variable struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

// MathSpec defines the desired state of Math
type MathSpec struct {
	Expression string     `json:"expression"`
	Variables  []Variable `json:"variables"`
}

// MathStatus defines the observed state of Math
type MathStatus struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Math is the Schema for the maths API
type Math struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MathSpec   `json:"spec,omitempty"`
	Status MathStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MathList contains a list of Math
type MathList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Math `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Math{}, &MathList{})
}
