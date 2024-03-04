package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +k8s:openapi-gen=true
// +resource:path=openhydrausers,strategy=openhydraStrategy,shortname=ohuser
// OpenHydraUser is the Schema for the OpenHydraUser API
type OpenHydraUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              OpenHydraUserSpec   `json:"spec,omitempty"`
	Status            OpenHydraUserStatus `json:"status,omitempty"`
}

// OpenHydraUserSpecUserStatus defines the observed state of Device of cluster
type OpenHydraUserStatus struct {
}

type OpenHydraUserSpec struct {
	ChineseName string `json:"chineseName,omitempty"`
	Description string `json:"description,omitempty"`
	Password    string `json:"password"`
	Email       string `json:"email,omitempty"`
	Role        int    `json:"role"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type OpenHydraUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenHydraUser `json:"items"`
}
