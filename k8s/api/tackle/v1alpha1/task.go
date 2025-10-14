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
	"encoding/json"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TaskSpec defines the desired state the resource.
type TaskSpec struct {
	// Priority defines the task priority (0-n).
	Priority int `json:"priority,omitempty"`
	// Dependencies defines a list of task names on which this task depends.
	Dependencies []string `json:"dependencies,omitempty"`
	// Data object passed to the addon.
	Data runtime.RawExtension `json:"data,omitempty"`
}

// TaskStatus defines the observed state the resource.
type TaskStatus struct {
	// The most recent generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// Task defines a hub task.
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
type Task struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	// Spec defines the desired state the resource.
	Spec TaskSpec `json:"spec,omitempty"`
	// Status defines the observed state the resource.
	Status TaskStatus `json:"status,omitempty"`
}

// Deps returns dependencies.
func (t *Task) Deps() []string {
	return t.Spec.Dependencies
}

// Data returns the task Data as map[string]any.
func (r *Task) Data() (mp map[string]any) {
	b := r.Spec.Data.Raw
	if b == nil {
		return
	}
	_ = json.Unmarshal(b, &mp)
	return
}

// TaskList is a list of Task.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TaskList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []Task `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Task{}, &TaskList{})
}
