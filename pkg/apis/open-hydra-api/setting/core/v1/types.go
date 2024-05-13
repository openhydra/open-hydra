package v1

import (
	"open-hydra/pkg/open-hydra/apis"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +k8s:openapi-gen=true
// +resource:path=settings,strategy=SettingStrategy,shortname=st
// Setting is the Schema for the Dataset API
type Setting struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SettingSpec   `json:"spec,omitempty"`
	Status            SettingStatus `json:"status,omitempty"`
}

// SettingStatus defines the observed state of Device of cluster
type SettingStatus struct {
}

type SettingSpec struct {
	DefaultGpuPerDevice uint8           `json:"default_gpu_per_device" yaml:"defaultGpuPerDevice"`
	PluginList          apis.PluginList `json:"plugin_list,omitempty" yaml:"pluginList,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type SettingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SettingSpec `json:"items"`
}
