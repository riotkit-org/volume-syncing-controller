package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=pfss

// PodFilesystemSync represents a filesystem/volume synchronization specification for given Pod
type PodFilesystemSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PodFilesystemSyncSpec `json:"spec"`
}

type PodFilesystemSyncSpec struct {
	PodSelector   PodSelector              `json:"pod_selector"`
	LocalPath     string                   `json:"local_path"`
	RemotePath    string                   `json:"remote_path"`
	Schedule      string                   `json:"schedule,omitempty"`
	Env           PodEnvironment           `json:"env,omitempty"`
	EnvFromSecret PodEnvironmentFromSecret `json:"env_from_secret"`
}

type PodSelector map[string]string
type PodEnvironment map[string]string
type PodEnvironmentFromSecret struct {
	Ref string `json:"ref"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// PodFilesystemSyncList represents a list
type PodFilesystemSyncList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodFilesystemSync `json:"items"`
}
