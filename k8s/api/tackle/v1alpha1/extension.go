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
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ExtensionSpec defines the desired state of the resource.
type ExtensionSpec struct {
	// Addon (name) declares addon compatibility.
	Addon string `json:"addon"`
	// Container defines the extension container.
	Container core.Container `json:"container"`
	// Selector defines criteria to be included in the addon pod.
	Selector string `json:"selector,omitempty"`
	// Metadata details.
	Metadata runtime.RawExtension `json:"metadata,omitempty"`
}

// ExtensionStatus defines the observed state of the resource.
type ExtensionStatus struct {
	// The most recent generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// Extension defines an addon extension.
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
type Extension struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// pec defines the desired state of the resource.
	Spec ExtensionSpec `json:"spec"`
	// Status defines the observed state of the resource.
	Status ExtensionStatus `json:"status,omitempty"`
}

// ExtensionList is a list of Extension.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ExtensionList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []Extension `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Extension{}, &ExtensionList{})
}
