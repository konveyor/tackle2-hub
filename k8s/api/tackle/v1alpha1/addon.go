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

// AddonSpec defines the desired state of the resource.
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
	// Task declares task compatability.
	Task string `json:"task,omitempty"`
	// Selector defines criteria to be selected for a task.
	Selector string `json:"selector,omitempty"`
	// Container defines the addon container.
	Container core.Container `json:"container,omitempty"`
	// Metadata details.
	Metadata runtime.RawExtension `json:"metadata,omitempty"`
}

// AddonStatus defines the observed state of the resource.
type AddonStatus struct {
	// The most recent generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Resource conditions.
	Conditions []meta.Condition `json:"conditions,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
type Addon struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// Spec defines the desired state of the resource.
	Spec AddonSpec `json:"spec"`
	// Status defines the observed state of the resource.
	Status AddonStatus `json:"status,omitempty"`
}

// Reconciled returns true when the resource has been reconciled.
func (r *Addon) Reconciled() (b bool) {
	return r.Generation == r.Status.ObservedGeneration
}

// Ready returns true when resource has the ready condition.
func (r *Addon) Ready() (ready bool) {
	for _, cnd := range r.Status.Conditions {
		if cnd.Type == Ready.Type && cnd.Status == meta.ConditionTrue {
			ready = true
			break
		}
	}
	return
}

// Migrate specification as needed.
func (r *Addon) Migrate() (updated bool) {
	if r.Spec.Image != nil {
		if r.Spec.Container.Image == "" {
			r.Spec.Container.Image = *r.Spec.Image
		}
		r.Spec.Image = nil
		updated = true
	}
	if r.Spec.Resources != nil {
		if len(r.Spec.Container.Resources.Limits) == 0 {
			r.Spec.Container.Resources.Limits = (*r.Spec.Resources).Limits
		}
		if len(r.Spec.Container.Resources.Requests) == 0 {
			r.Spec.Container.Resources.Requests = (*r.Spec.Resources).Requests
		}
		r.Spec.Resources = nil
		updated = true
	}
	if r.Spec.ImagePullPolicy != nil {
		if r.Spec.Container.ImagePullPolicy == "" {
			r.Spec.Container.ImagePullPolicy = *r.Spec.ImagePullPolicy
		}
		r.Spec.ImagePullPolicy = nil
		updated = true
	}
	if r.Spec.Container.Name == "" {
		r.Spec.Container.Name = "addon"
		updated = true
	}
	return
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
