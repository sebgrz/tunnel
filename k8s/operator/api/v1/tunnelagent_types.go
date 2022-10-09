/*
Copyright 2022.

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

// TunnelAgentSpec defines the desired state of TunnelAgent
type TunnelAgentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Server is an address to a running tunnel server
	Server string `json:"server"`
	// Hostname determine correct proxy path
	Hostname string `json:"hostname"`
	// Type of connection: http or ws
	Type string `json:"type"`
}

// TunnelAgentStatus defines the observed state of TunnelAgent
type TunnelAgentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TunnelAgent is the Schema for the tunnelagents API
type TunnelAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TunnelAgentSpec   `json:"spec,omitempty"`
	Status TunnelAgentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TunnelAgentList contains a list of TunnelAgent
type TunnelAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TunnelAgent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TunnelAgent{}, &TunnelAgentList{})
}
