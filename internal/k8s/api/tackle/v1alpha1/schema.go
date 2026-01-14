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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// SchemaVersion defines each version of a schema.
type SchemaVersion struct {
	// Migration defines a yq query to migrate the document.
	Migration string `json:"migration,omitempty"`
	// Definition is the (jsd) json-schema definition.
	Definition runtime.RawExtension `json:"definition"`
}

// SchemaSpec defines the desired state of the resource.
type SchemaSpec struct {
	// Domain
	Domain string `json:"domain"`
	// Variant
	Variant string `json:"variant"`
	// Subject
	Subject string `json:"subject"`
	// Versions
	Versions []SchemaVersion `json:"versions"`
}

// SchemaStatus defines the observed state of the resource.
type SchemaStatus struct {
	// The most recent generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// Schema defines json document schemas.
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
type Schema struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            SchemaSpec   `json:"spec"`
	Status          SchemaStatus `json:"status,omitempty"`
}

// SchemaList is a list of Schema.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SchemaList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []Schema `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Schema{}, &SchemaList{})
}
