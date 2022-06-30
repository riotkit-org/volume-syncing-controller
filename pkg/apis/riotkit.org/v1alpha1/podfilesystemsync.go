package v1alpha1

import (
	"github.com/flosch/pongo2/v6"
	"github.com/pkg/errors"
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
// +kubebuilder:subresource:status
type PodFilesystemSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodFilesystemSyncSpec `json:"spec"`
	Status SynchronizationStatus `json:"status,omitempty"`
}

// NewPodFilesystemSync is making a new instance of a resource making sure that defaults will be respected
func NewPodFilesystemSync() PodFilesystemSync {
	return PodFilesystemSync{
		Spec: PodFilesystemSyncSpec{
			SyncOptions: SyncOptionsSpec{
				Schedule:                "@every 15m",
				Method:                  "scheduler",
				RestoreRemoteOnFirstRun: true,
				CleanUp: CleanUpSpec{
					Remote:      true,
					Local:       true,
					ForceRemote: true,
					ForceLocal:  true,
				},
				AllowedDirections: AllowedDirectionsSpec{
					ToRemote:   true,
					FromRemote: true,
				},
			},
			AutomaticEncryption: AutomaticEncryptionSpec{
				Enabled: false,
			},
			Debug: false,
		},
	}
}

type CleanUpSpec struct {
	// +kubebuilder:default:=true
	Remote bool `json:"remote,omitempty"`
	// +kubebuilder:default:=true
	Local bool `json:"local,omitempty"`
	// +kubebuilder:default:=false
	ForceRemote bool `json:"forceRemote,omitempty"`
	// +kubebuilder:default:=false
	ForceLocal bool `json:"forceLocal,omitempty"`
}

// AutomaticEncryptionSpec enables automatic encryption
type AutomaticEncryptionSpec struct {
	// +kubebuilder:default:=false
	Enabled    bool   `json:"enabled,omitempty"`
	SecretName string `json:"secretName"`
}

// AllowedDirectionsSpec is describing if we can upload or download files
type AllowedDirectionsSpec struct {
	// +kubebuilder:default:=true
	ToRemote bool `json:"toRemote,omitempty"`

	// +kubebuilder:default:=true
	FromRemote bool `json:"fromRemote,omitempty"`
}

// SyncOptionsSpec .spec.syncOptions, synchronization detailed options
type SyncOptionsSpec struct {
	// +kubebuilder:default:=@every 15m
	Schedule string `json:"schedule,omitempty"` // will default to every 15 minutes
	// +kubebuilder:default:=scheduler
	Method               ChangesWatchingMethod `json:"method,omitempty"` // scheduler or fsnotify
	MaxOneSyncPerMinutes string                `json:"maxOneSyncPerMinutes,omitempty"`
	Permissions          PermissionsSpec       `json:"permissions,omitempty"`

	// +kubebuilder:default:=true
	RestoreRemoteOnFirstRun bool `json:"restoreRemoteOnFirstRun,omitempty"`

	CleanUp           CleanUpSpec           `json:"cleanUp,omitempty"`
	AllowedDirections AllowedDirectionsSpec `json:"allowedDirections,omitempty"`
}

// PermissionsSpec defines permissions to files inside Pod, to be able to run as non-root
type PermissionsSpec struct {
	UID string `json:"uid,omitempty"`
	GID string `json:"gid,omitempty"`
}

// PodFilesystemSyncSpec represents .spec
type PodFilesystemSyncSpec struct {
	PodSelector *metav1.LabelSelector `json:"podSelector"`

	LocalPath  string `json:"localPath"`
	RemotePath string `json:"remotePath"`

	SyncOptions SyncOptionsSpec `json:"syncOptions"`

	// use environment to configure remotes and encryption
	// values can contain Go-Template syntax e.g. {{ .pod.metadata.labels["some-label"] }}
	Env            PodEnvironment            `json:"env,omitempty"`
	EnvFromSecrets PodEnvironmentFromSecrets `json:"envFromSecrets,omitempty"`

	// automatic encryption is creating a `kind: Secret` if not exists and configuring encryption automatically
	AutomaticEncryption AutomaticEncryptionSpec `json:"automaticEncryption,omitempty"`
	// +kubebuilder:default:=false
	Debug bool `json:"debug,omitempty"`
}

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

type SynchronizationStatus struct {
	Locations []LocationStatus `json:"locations"`
}

type LocationStatus struct {
	Directory            string `json:"directory"`
	SynchronizedToRemote bool   `json:"synchronizedToRemote"`
}

// ResolveDirectoryForPod is building remote path that will be used for given Pod. Depending on the configuration the `.Spec.RemotePath` may be a JINJA2 template
//                        and this method allows to use syntax like {% pod.metadata.labels["directoryName"] %}
func (in *PodFilesystemSync) ResolveDirectoryForPod(pod *v1.Pod) (string, error) {
	tmpl, templateErr := pongo2.FromString(in.Spec.RemotePath)
	if templateErr != nil {
		return "", errors.Wrapf(templateErr, "Cannot build remote path using template '%v' - parse error", in.Spec.RemotePath)
	}
	out, err := tmpl.Execute(pongo2.Context{"pod": pod})
	if err != nil {
		return "", errors.Wrapf(err, "Cannot build remote path using template '%v' - evaluation error", in.Spec.RemotePath)
	}
	return out, nil
}

func (in *PodFilesystemSync) findLocation(path string) (bool, *LocationStatus) {
	for _, location := range in.Status.Locations {
		if location.Directory == path {
			return true, &location
		}
	}
	return false, &LocationStatus{Directory: path, SynchronizedToRemote: false}
}

// WasAlreadySynchronized tells if Pod's filesystem was already at least one time synchronized
func (in *PodFilesystemSync) WasAlreadySynchronized(pod *v1.Pod) (bool, error) {
	directory, err := in.ResolveDirectoryForPod(pod)
	if err != nil {
		return false, err
	}
	existing, location := in.findLocation(directory)
	wasAlreadySynchronized := existing && location.SynchronizedToRemote

	return wasAlreadySynchronized, nil
}

// ShouldRestoreFilesFromRemote decides if files could be restored from remote
func (in *PodFilesystemSync) ShouldRestoreFilesFromRemote(pod *v1.Pod) (bool, error) {
	wasAlreadySynchronized, err := in.WasAlreadySynchronized(pod)
	if err != nil {
		return false, err
	}
	if !in.Spec.SyncOptions.AllowedDirections.FromRemote {
		logrus.Debugf("FromRemote direction disallowed for PodFilesystemSync '%s'", in.GetName())
		return false, nil
	}
	return wasAlreadySynchronized || (!wasAlreadySynchronized && in.Spec.SyncOptions.RestoreRemoteOnFirstRun), nil
}

func (in *PodFilesystemSync) ShouldSynchronizeToRemote() bool {
	return in.Spec.SyncOptions.AllowedDirections.ToRemote
}

// ClaimDirectoryByPod mark target directory claimed by Pod as synchronized
func (in *PodFilesystemSync) ClaimDirectoryByPod(pod *v1.Pod) error {
	directory, err := in.ResolveDirectoryForPod(pod)
	if err != nil {
		return err
	}

	existing, location := in.findLocation(directory)
	location.SynchronizedToRemote = true

	if !existing {
		logrus.Debug("Adding new location to status")
		in.Status.Locations = append(in.Status.Locations, *location)
	}
	return nil
}
