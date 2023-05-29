/*
Copyright 2023.

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

// MinecraftServerSpec defines the desired state of MinecraftServer
type MinecraftServerSpec struct {
	Active bool `json:"active"`

	Server  ServerConfig  `json:"server"`
	Storage StorageConfig `json:"storage"`
}

// MinecraftServerStatus defines the observed state of MinecraftServer
type MinecraftServerStatus struct {
	State    ServerState    `json:"state"`
	Runner   RunnerStatus   `json:"runner"`
	Storage  StorageStatus  `json:"storage"`
	Uploader UploaderStatus `json:"uploader"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state"
//+kubebuilder:printcolumn:name="Runner",type="string",JSONPath=".status.runner.podRef.name"
//+kubebuilder:printcolumn:name="Uploader",type="string",JSONPath=".status.uploader.podRef.name"
//+kubebuilder:printcolumn:name="Storage",type="string",JSONPath=".status.storage.pvcRef.name"

// MinecraftServer is the Schema for the minecraftservers API
type MinecraftServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MinecraftServerSpec   `json:"spec,omitempty"`
	Status MinecraftServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MinecraftServerList contains a list of MinecraftServer
type MinecraftServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MinecraftServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MinecraftServer{}, &MinecraftServerList{})
}
