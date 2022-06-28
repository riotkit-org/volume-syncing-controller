package v1alpha1

import (
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// +kubebuilder:validation:Enum=scheduler;fsnotify
type ChangesWatchingMethod string

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

type CleanUpSpec struct {
	Remote      bool `json:"remote,omitempty"`
	Local       bool `json:"local,omitempty"`
	ForceRemote bool `json:"forceRemote,omitempty"`
	ForceLocal  bool `json:"forceLocal,omitempty"`
}

type AutomaticEncryptionSpec struct {
	Enabled    bool   `json:"enabled,omitempty"`
	SecretName string `json:"secretName"`
}

type SyncOptionsSpec struct {
	Schedule             string                `json:"schedule,omitempty"` // will default to every 15 minutes
	Method               ChangesWatchingMethod `json:"method,omitempty"`   // scheduler or fsnotify
	MaxOneSyncPerMinutes string                `json:"maxOneSyncPerMinutes,omitempty"`
	Permissions          PermissionsSpec       `json:"permissions,omitempty"`
}

type PermissionsSpec struct {
	UID string `json:"uid,omitempty"`
	GID string `json:"gid,omitempty"`
}

type PodFilesystemSyncSpec struct {
	PodSelector *metav1.LabelSelector `json:"podSelector"`

	LocalPath  string `json:"localPath"`
	RemotePath string `json:"remotePath"`

	SyncOptions SyncOptionsSpec `json:"syncOptions,omitempty"`

	// use environment to configure remotes and encryption
	// values can contain Go-Template syntax e.g. {{ .pod.metadata.labels["some-label"] }}
	Env            PodEnvironment            `json:"env,omitempty"`
	EnvFromSecrets PodEnvironmentFromSecrets `json:"envFromSecrets,omitempty"`

	// automatic encryption is creating a `kind: Secret` if not exists and configuring encryption automatically
	AutomaticEncryption AutomaticEncryptionSpec `json:"automaticEncryption,omitempty"`
	Debug               bool                    `json:"debug,omitempty"`
	CleanUp             CleanUpSpec             `json:"cleanUp,omitempty"`
}

type PodSelector map[string]string
type PodEnvironment map[string]string
type PodEnvironmentFromSecrets []v1.SecretReference

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// PodFilesystemSyncList represents a list
type PodFilesystemSyncList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodFilesystemSync `json:"items"`
}

// getPodSelector dynamically constructs a labels.Selector if needed
func (in *PodFilesystemSyncSpec) getPodSelector() labels.Selector {
	podSelector, err := metav1.LabelSelectorAsSelector(in.PodSelector)
	if err != nil {
		logrus.Errorf("Invalid podSelector syntax: '%v'. Selector: '%v'", err, in.PodSelector.String())
	}
	return podSelector
}

// IsPodMatching is `kind: Pod` matching .spec.podSelector of `kind: PodFilesystemSync`?
func (in *PodFilesystemSync) IsPodMatching(pod *v1.Pod) bool {
	return in.Spec.getPodSelector().Matches(labels.Set(pod.Labels))
}
