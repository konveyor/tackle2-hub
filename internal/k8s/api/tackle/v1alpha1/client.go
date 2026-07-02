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
)

// IdpClientSpec defines the desired state of IdpClient.
type IdpClientSpec struct {
	// ID is the database ID for the seeded client.
	// Must be less than 1000 (reserved range for seeded clients).
	// +kubebuilder:validation:Maximum=999
	// +kubebuilder:validation:Minimum=1
	ID uint `json:"id"`
	// ClientId is the OAuth client identifier (e.g., "web-ui", "kantra").
	// This is used as the natural key for reconciliation.
	ClientId string `json:"clientId"`
	// ClientSecret references a Kubernetes Secret containing the OAuth client secret.
	// The Secret must have a key named "clientSecret".
	// This is optional - public clients (e.g., native apps) may not require a secret.
	// +optional
	ClientSecret *core.ObjectReference `json:"clientSecret,omitempty"`
	// ApplicationType is the OAuth application type (e.g., "web", "native").
	ApplicationType string `json:"applicationType"`
	// Grants are the OAuth grant types supported by this client.
	Grants []string `json:"grants"`
	// RedirectURIs are the redirect URIs for OAuth flows.
	// +optional
	RedirectURIs []string `json:"redirectURIs"`
	// Scopes are the OAuth scopes requested by this client.
	Scopes []string `json:"scopes"`
}

// IdpClientStatus defines the observed state of IdpClient.
type IdpClientStatus struct {
	// The most recent generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Resource conditions.
	Conditions []meta.Condition `json:"conditions,omitempty"`
}

// IdpClient defines an OIDC client configuration.
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=client
type IdpClient struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// Spec defines the desired state of the resource.
	Spec IdpClientSpec `json:"spec"`
	// Status defines the observed state of the resource.
	Status IdpClientStatus `json:"status,omitempty"`
}

// Reconciled returns true when the resource has been reconciled.
func (r *IdpClient) Reconciled() (b bool) {
	return r.Generation == r.Status.ObservedGeneration
}

// Ready returns true when resource has the ready condition.
func (r *IdpClient) Ready() (ready bool) {
	for _, cnd := range r.Status.Conditions {
		if cnd.Type == Ready.Type && cnd.Status == meta.ConditionTrue {
			ready = true
			break
		}
	}
	return
}

// IdpClientList contains a list of IdpClient.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type IdpClientList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []IdpClient `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IdpClient{}, &IdpClientList{})
}
