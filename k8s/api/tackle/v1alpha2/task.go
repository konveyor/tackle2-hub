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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TaskSpec defines the desired state of aTask.
type TaskSpec struct {
	// Priority defines the task priority (0-n).
	Priority int `json:"priority,omitempty"`
	// Dependencies defines a list of task names on which this task depends.
	Dependencies []string `json:"dependencies,omitempty"`
	// Data object passed to the addon.
	Data runtime.RawExtension `json:"data,omitempty"`
}

// TaskStatus defines the observed state of a Task.
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
	// Spec defines the desired state of Task.
	Spec TaskSpec `json:"spec,omitempty"`
	// Status defines the observed state of Task
	Status TaskStatus `json:"status,omitempty"`
}

// HasDep return true if the task has the dependency.
func (r *Task) HasDep(name string) (found bool) {
	for i := range r.Spec.Dependencies {
		n := r.Spec.Dependencies[i]
		if n == name {
			found = true
			break
		}
	}
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
