package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +k8s:openapi-gen=true
// +resource:path=sumups,strategy=sumupStrategy,shortname=xsu
// SumUp is the Schema for the cluster summary API
type SumUp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SumUpSpec   `json:"spec,omitempty"`
	Status            SumUpStatus `json:"status,omitempty"`
}

// SumUpStatus defines the observed state of Device of cluster
type SumUpStatus struct {
}

type SumUpSpec struct {
	PodAllocatable      int                         `json:"podAllocatable"`
	PodAllocated        int                         `json:"podAllocated"`
	GpuAllocatable      string                      `json:"gpuAllocatable"`
	GpuAllocated        string                      `json:"gpuAllocated"`
	DefaultCpuPerDevice string                      `json:"defaultCpuPerDevice"`
	DefaultRamPerDevice string                      `json:"defaultRamPerDevice"`
	DefaultGpuPerDevice uint8                       `json:"defaultGpuPerDevice"`
	TotalLine           uint16                      `json:"totalLine"`
	GpuResourceSumUp    map[string]GpuResourceSumUp `json:"gpuResourceSumUp"`
}

type GpuResourceSumUp struct {
	Allocated   int64 `json:"allocated"`
	Allocatable int64 `json:"allocatable"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type SumUpList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SumUp `json:"items"`
}
