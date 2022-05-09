/*
Copyright 2019 Red Hat Inc.

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

package v1alpha1

import (
	"github.com/konveyor/controller/pkg/condition"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//
// Mount specification.
type Mount struct {
	Name  string `json:"name"`
	Claim string `json:"claim"`
}

//
// AddonSpec defines the desired state of Addon
type AddonSpec struct {
	// Addon fqin.
	Image string `json:"image"`
	// Resource requirements.
	Resources core.ResourceRequirements `json:"resources,omitempty"`
	// Mounts optional.
	Mounts []Mount `json:"mounts,omitempty"`
}

//
// AddonStatus defines the observed state of Addon
type AddonStatus struct {
	//
	// Conditions.
	condition.Conditions `json:"conditions"`
	// The most recent generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type=string,JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="CONNECTED",type=string,JSONPath=".status.conditions[?(@.type=='ConnectionTestSucceeded')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type Addon struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            AddonSpec   `json:"spec,omitempty"`
	Status          AddonStatus `json:"status,omitempty"`
}

//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AddonList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []Addon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Addon{}, &AddonList{})
}
