package v1

import "k8s.io/apimachinery/pkg/api/resource"

// ServerConfig defines the configuration for the Pod to be run
type ServerConfig struct {
	ID      string `json:"id"`
	Jarfile string `json:"jarfile"`

	HeapMax  string `json:"heapMax"`
	HeapInit string `json:"heapInit"`
}

// StorageConfig defines the configuration for the PersistentVolumeClaim
// to be provisioned
type StorageConfig struct {
	Capacity resource.Quantity `json:"capacity"`
}
