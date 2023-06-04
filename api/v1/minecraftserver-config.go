package v1

import (
	"k8s.io/apimachinery/pkg/api/resource"

	corev1 "k8s.io/api/core/v1"
)

// ServerConfig defines the configuration for the Pod to be run
type ServerConfig struct {
	ID string `json:"id"`

	Image string          `json:"image"`
	Env   []corev1.EnvVar `json:"env"`
}

// StorageConfig defines the configuration for the PersistentVolumeClaim
// to be provisioned
type StorageConfig struct {
	Capacity resource.Quantity `json:"capacity"`
}
