package v1alpha1

import (
	// corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type VirtualMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualMachineSpec   `json:"spec"`
	Status VirtualMachineStatus `json:"status"`
}



type VirtualMachineSpec struct {
	// Resource corev1.ResourceList `json:"resource"`
	// Cpu    int64 `json:"cpu"`
	Cpu    resource.Quantity `json:"cpu"`
	Memory resource.Quantity `json:"memory"`
}

type VirtualMachinePhase string

const (
	VirtualMachineNone        VirtualMachinePhase = ""
	VirtualMachineCreating    VirtualMachinePhase = "Creating"
	VirtualMachineActive      VirtualMachinePhase = "Active"
	VirtualMachineFailed      VirtualMachinePhase = "Failed"
	VirtualMachineTerminating VirtualMachinePhase = "Terminating"
	VirtualMachineUnknown     VirtualMachinePhase = "Unknown"
	// VirtualMachinePending     VirtualMachinePhase = "Pending"
)

type ResourceUsage struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
}

type ServerStatus struct {
	ID    string        `json:"id"`
	State string        `json:"state"`
	Usage ResourceUsage `json:"usage"`
}

type VirtualMachineStatus struct {
	Phase          VirtualMachinePhase `json:"phase"`
	Reason         string              `json:"reason,omitempty"`
	Server         ServerStatus        `json:"server,omitempty"`
	LastUpdateTime metav1.Time         `json:"lastUpdateTime"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type VirtualMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []VirtualMachine `json:"items"`
}
