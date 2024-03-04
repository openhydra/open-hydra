package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +k8s:openapi-gen=true
// +resource:path=datasets,strategy=DatasetStrategy,shortname=dst
// Dataset is the Schema for the Dataset API
type Dataset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DatasetSpec   `json:"spec,omitempty"`
	Status            DatasetStatus `json:"status,omitempty"`
}

// DatasetStatus defines the observed state of Device of cluster
type DatasetStatus struct {
}

type DatasetSpec struct {
	Description string      `json:"description,omitempty"`
	LastUpdate  metav1.Time `json:"lastUpdate"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type DatasetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dataset `json:"items"`
}
