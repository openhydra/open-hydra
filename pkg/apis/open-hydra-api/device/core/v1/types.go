package v1

import (
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +k8s:openapi-gen=true
// +resource:path=devices,strategy=DeviceStrategy,shortname=dev
// Device is the Schema for the Device API
type Device struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DeviceSpec   `json:"spec,omitempty"`
	Status            DeviceStatus `json:"status,omitempty"`
}

// DeviceStatus defines the observed state of Device of cluster
type DeviceStatus struct {
}

type DeviceSpec struct {
	DeviceName        string           `json:"deviceName,omitempty"`
	DeviceNamespace   string           `json:"deviceNamespace,omitempty"`
	DeviceType        string           `json:"deviceType,omitempty"`
	DeviceIP          string           `json:"deviceIP,omitempty"`
	DeviceCpu         string           `json:"deviceCpu,omitempty"`
	DeviceRam         string           `json:"deviceRam,omitempty"`
	DeviceGpu         uint8            `json:"deviceGpu,omitempty"`
	DeviceStatus      string           `json:"deviceStatus,omitempty"`
	GpuDriver         string           `json:"gpuDriver,omitempty"`
	OpenHydraUsername string           `json:"openHydraUsername,omitempty"`
	Role              int              `json:"role,omitempty"`
	ChineseName       string           `json:"chineseName,omitempty"`
	LineNo            string           `json:"lineNo,omitempty"`
	UsePublicDataSet  bool             `json:"usePublicDataSet,omitempty"`
	SandboxURLs       string           `json:"sandboxURLs,omitempty"`
	SandboxName       string           `json:"sandboxName,omitempty"`
	Affinity          *coreV1.Affinity `json:"affinity,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type DeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Device `json:"items"`
}
