package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServerState string

const (
	Initializing ServerState = "Initializing"
	Idling       ServerState = "Idling"
	Activating   ServerState = "Activating"
	Running      ServerState = "Running"
	Deactivating ServerState = "Deactivating"
	BackingUp    ServerState = "Backing Up"
	CleaningUp   ServerState = "Cleaning Up"
	Deleting     ServerState = "Deleting"
	Error        ServerState = "Error"
)

type RunnerStatus struct {
	PodRef *corev1.ObjectReference `json:"podRef,omitempty"`

	PodIP string `json:"podIP,omitempty"`
}

type UploaderStatus struct {
	PodRef *corev1.ObjectReference `json:"podRef,omitempty"`
}

type StorageStatus struct {
	PVCRef *corev1.ObjectReference `json:"pvcRef,omitempty"`

	Initialized  string       `json:"initialized,omitempty"`
	LatestBackup *metav1.Time `json:"latestBackup,omitempty"`
}
