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

// RoleMapping defines pattern matching for LDAP groups to roles.
type RoleMapping struct {
	// Any patterns (OR condition).
	Any []string `json:"any,omitempty"`
	// And patterns (AND condition).
	And []string `json:"and,omitempty"`
	// Role name to assign when matched.
	Roles []string `json:"roles"`
}

// LdapProviderSpec defines the desired state of the resource.
type LdapProviderSpec struct {
	// Provider name.
	Name string `json:"name"`
	// LDAP kind (ACTIVEDIRECTORY, AD, or blank for standard LDAP).
	// +optional
	Kind string `json:"kind,omitempty"`
	// LDAP server URL (e.g., ldap://ldap.example.com:389).
	URL string `json:"url"`
	// Base DN for LDAP searches (e.g., dc=example,dc=com).
	BaseDN string `json:"baseDN"`
	// Service account bind DN for LDAP authentication.
	BindDN string `json:"bindDN"`
	// Password reference for service account authentication.
	Password *core.ObjectReference `json:"password"`
	// Custom user search filter (optional, defaults based on Kind).
	// +optional
	UserFilter string `json:"userFilter,omitempty"`
	// Custom group search filter (optional, defaults based on Kind).
	// +optional
	GroupFilter string `json:"groupFilter,omitempty"`
	// Use memberOf attribute for group membership (faster if available).
	// +optional
	HasMemberOf bool `json:"hasMemberOf,omitempty"`
	// Role mappings from LDAP groups to application roles.
	RoleMappings []RoleMapping `json:"roleMappings"`
	// TLS connection settings.
	// +optional
	TLS TLS `json:"tls,omitempty"`
}

// LdapProviderStatus defines the observed state of the resource.
type LdapProviderStatus struct {
	// The most recent generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Resource conditions.
	Conditions []meta.Condition `json:"conditions,omitempty"`
}

// LdapProvider defines LDAP authentication and authorization settings.
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ldap
type LdapProvider struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// Spec defines the desired state of the resource.
	Spec LdapProviderSpec `json:"spec"`
	// Status defines the observed state of the resource.
	Status LdapProviderStatus `json:"status,omitempty"`
}

// Reconciled returns true when the resource has been reconciled.
func (r *LdapProvider) Reconciled() (b bool) {
	return r.Generation == r.Status.ObservedGeneration
}

// Ready returns true when resource has the ready condition.
func (r *LdapProvider) Ready() (ready bool) {
	for _, cnd := range r.Status.Conditions {
		if cnd.Type == Ready.Type && cnd.Status == meta.ConditionTrue {
			ready = true
			break
		}
	}
	return
}

// LdapProviderList is a list of LdapProvider.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type LdapProviderList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []LdapProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LdapProvider{}, &LdapProviderList{})
}
