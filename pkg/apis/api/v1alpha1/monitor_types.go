package v1alpha1

import (
	"crypto/sha1"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MonitorSpec defines the desired state of Monitor
type MonitorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	TargetSite string `json:"targetSite"`
	Data       string `json:"data"`
}

func (in *MonitorSpec) Checksum() string {
	sum := sha1.Sum([]byte(in.Data))
	return fmt.Sprintf("%x", sum)
}

type MonitorState string

const (
	MonitorStateCreated = "created"
	MonitorStateDeleted = "deleted"
	MonitorStateError   = "error"
)

// MonitorStatus defines the observed state of Monitor
type MonitorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	State               MonitorState `json:"state"`
	MonitorID           int          `json:"ID"`
	LastAppliedChecksum string       `json:"lastAppliedChecksum"`
	ErrorMessage        string       `json:"errorMessage"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Monitor is the Schema for the monitors API
// +k8s:openapi-gen=true
type Monitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MonitorSpec   `json:"spec,omitempty"`
	Status MonitorStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MonitorList contains a list of Monitor
type MonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Monitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Monitor{}, &MonitorList{})
}
