package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InfinispanSpec defines all Infinispan specific configuration
type InfinispanSpec struct {
}

// RedisSpec defines all Redis specific configuration
type RedisSpec struct {
}

// CacheSpec defines the desired state of Cache
type CacheSpec struct {
	Infinispan *InfinispanSpec `json:"infinispan,omitempty"`
	Redis      *RedisSpec      `json:"redis,omitempty"`
}

// CacheStatus defines the observed state of Cache
type CacheStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Cache is the Schema for the caches API
type Cache struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CacheSpec   `json:"spec,omitempty"`
	Status CacheStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CacheList contains a list of Cache
type CacheList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cache `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cache{}, &CacheList{})
}
