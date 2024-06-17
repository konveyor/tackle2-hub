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

package v1alpha2

import (
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// AddonSpec defines the desired state of an Addon.
type AddonSpec struct {
	// Deprecated: Addon is deprecated.
	// +kubebuilder:validation:Optional
	Image *string `json:"image,omitempty"`
	// Deprecated: ImagePullPolicy is deprecated.
	// +kubebuilder:validation:Optional
	ImagePullPolicy *core.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Deprecated: Resources is deprecated.
	// +kubebuilder:validation:Optional
	Resources *core.ResourceRequirements `json:"resources,omitempty"`
	//
	// Task declares task (kind) compatibility.
	Task string `json:"task,omitempty"`
	// Selector defines criteria to be selected for a task.
	Selector string `json:"selector,omitempty"`
	// Container defines the addon container.
	Container core.Container `json:"container"`
	// Metadata details.
	Metadata runtime.RawExtension `json:"metadata,omitempty"`
}

// AddonStatus defines the observed state of an Addon.
type AddonStatus struct {
	// The most recent generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
type Addon struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// Spec defines the desired state of an Addon.
	Spec AddonSpec `json:"spec,omitempty"`
	// Status defines the observed state of an Addon.
	Status AddonStatus `json:"status,omitempty"`
}

// AddonList is a list of Addon.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AddonList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []Addon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Addon{}, &AddonList{})
}
