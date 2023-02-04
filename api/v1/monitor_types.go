/*
Copyright 2023 wangan.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MonitorSpec defines the desired state of Monitor
type MonitorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Monitor. Edit monitor_types.go to remove/update
	UpdateInterval int64 `json:"updateInterval,omitempty"`
}

// MonitorStatus defines the observed state of Monitor
type MonitorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	CardList       CardList     `json:"cardList,omitempty"`
	CardNumber     uint         `json:"cardNumber,omitempty"`
	UpdateTime     *metav1.Time `json:"updateTime,omitempty"`
	TotalMemorySum uint64       `json:"totalMemorySum,omitempty"`
	FreeMemorySum  uint64       `json:"freeMemorySum,omitempty"`
}

type CardList []Card

func (in CardList) Len() int {
	return len(in)
}

func (in CardList) Less(i, j int) bool {
	return in[i].ID < in[j].ID
}

func (in CardList) Swap(i, j int) {
	in[i], in[j] = in[j], in[i]
}

type Card struct {
	ID          uint   `json:"id"`
	Health      string `json:"health,omitempty"`
	Model       string `json:"model,omitempty"`
	Power       uint   `json:"power,omitempty"`
	TotalMemory uint64 `json:"totalMemory,omitempty"`
	Clock       uint   `json:"clock,omitempty"`
	FreeMemory  uint64 `json:"freeMemory,omitempty"`
	Core        uint   `json:"core,omitempty"`
	Bandwidth   uint   `json:"bandwidth,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster

// Monitor is the Schema for the monitors API
type Monitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MonitorSpec   `json:"spec,omitempty"`
	Status MonitorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MonitorList contains a list of Monitor
type MonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Monitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Monitor{}, &MonitorList{})
}
