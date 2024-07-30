package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +k8s:openapi-gen=true
// +resource:path=courses,strategy=CourseStrategy,shortname=dst
// Course is the Schema for the Course API
type Course struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CourseSpec   `json:"spec,omitempty"`
	Status            CourseStatus `json:"status,omitempty"`
}

// CourseStatus defines the observed state of Device of cluster
type CourseStatus struct {
}

type CourseSpec struct {
	CreatedBy   string      `json:"createdBy,omitempty"`
	Description string      `json:"description,omitempty"`
	LastUpdate  metav1.Time `json:"lastUpdate"`
	Level       int         `json:"level,omitempty"`
	SandboxName string      `json:"sandboxName,omitempty"`
	Size        int64       `json:"size,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type CourseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Course `json:"items"`
}
