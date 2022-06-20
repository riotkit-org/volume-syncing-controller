package context

import (
	"github.com/riotkit-org/volume-syncing-operator/pkg/apis/riotkit.org/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type SynchronizationParameters struct {
	LocalPath  string
	RemotePath string

	// Sync Options
	SyncSchedule             string
	SyncMethod               string
	SyncMaxOneSyncPerMinutes string

	// Environments already mixed together
	Env        map[string]string
	EnvSecrets []string // will be rendered as `envFrom`

	Debug bool

	// Clean up (sync vs copy)
	CleanUpRemote      bool
	ForceCleanUpRemote bool
	CleanUpLocal       bool
	ForceCleanUpLocal  bool

	// Optional owner UID & group GID
	Owner string
	Group string
}

// CreateCommandlineArgumentsForInitContainer is creating commandline arguments for volume-syncing-operator remote-to-local-sync
func (p *SynchronizationParameters) CreateCommandlineArgumentsForInitContainer() []string {
	args := []string{
		"--src", p.RemotePath,
		"--dst", p.LocalPath,
	}

	if p.SyncSchedule != "" && (p.SyncMethod == "scheduler" || p.SyncMethod == "") {
		args = append(args, "--schedule", p.SyncSchedule)
	}
	if p.SyncMethod == "fsnotify" {
		args = append(args, "--fsnotify", string(p.SyncMaxOneSyncPerMinutes))
	}
	if p.Debug {
		args = append(args, "--verbose")
	}
	if !p.CleanUpLocal {
		args = append(args, "--no-delete")
	}
	if p.ForceCleanUpLocal {
		args = append(args, "--force-delete-local-dir")
	}

	return args
}

// CreateCommandlineArgumentsForSideCar is creating commandline args for volume-syncing-operator sync-to-remote
func (p *SynchronizationParameters) CreateCommandlineArgumentsForSideCar() []string {
	args := []string{
		"--src", p.LocalPath,
		"--dst", p.RemotePath,
	}

	if p.SyncSchedule != "" && (p.SyncMethod == "scheduler" || p.SyncMethod == "") {
		args = append(args, "--schedule", p.SyncSchedule)
	}
	if p.SyncMethod == "fsnotify" {
		args = append(args, "--fsnotify", string(p.SyncMaxOneSyncPerMinutes))
	}
	if !p.CleanUpRemote {
		args = append(args, "--no-delete")
	}
	if p.ForceCleanUpRemote {
		args = append(args, "--force-even-if-remote-would-be-cleared")
	}
	if p.Debug {
		args = append(args, "--verbose")
	}

	return args
}

// NewSynchronizationParameters constructs a unified parameters context mapped from CRD into a format used by Mutator
// the "env" parameter should be already a merged list of environment variables, with resolved `kind: Secret` objects into environment variables
func NewSynchronizationParameters(pod *corev1.Pod, syncDefinition *v1alpha1.PodFilesystemSync, env map[string]string) SynchronizationParameters {
	uid := syncDefinition.Spec.SyncOptions.Permissions.UID
	gid := syncDefinition.Spec.SyncOptions.Permissions.GID

	// allow to override UID and GID permissions by labels
	if val, exists := pod.Annotations[AnnotationUid]; exists {
		uid = val
	}
	if val, exists := pod.Annotations[AnnotationGid]; exists {
		gid = val
	}

	// `kind: Secret`
	// convert map to list to be used in envFrom
	var envSecrets []string
	for _, secret := range syncDefinition.Spec.EnvFromSecrets {
		envSecrets = append(envSecrets, secret.Name)
	}

	return SynchronizationParameters{
		LocalPath:                syncDefinition.Spec.LocalPath,
		RemotePath:               syncDefinition.Spec.RemotePath,
		SyncSchedule:             syncDefinition.Spec.SyncOptions.Schedule,
		SyncMethod:               string(syncDefinition.Spec.SyncOptions.Method),
		SyncMaxOneSyncPerMinutes: syncDefinition.Spec.SyncOptions.MaxOneSyncPerMinutes,
		Env:                      env,
		EnvSecrets:               envSecrets,
		Debug:                    syncDefinition.Spec.Debug,

		CleanUpRemote:      syncDefinition.Spec.CleanUp.Remote,
		ForceCleanUpRemote: syncDefinition.Spec.CleanUp.ForceRemote,
		CleanUpLocal:       syncDefinition.Spec.CleanUp.Local,
		ForceCleanUpLocal:  syncDefinition.Spec.CleanUp.ForceLocal,

		Owner: uid,
		Group: gid,
	}
}