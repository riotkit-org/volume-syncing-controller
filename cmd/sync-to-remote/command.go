package sync_to_remote

import "github.com/riotkit-org/volume-syncer/pkg/rclone"

type SyncToRemoteCommand struct {
	configPath   string
	renderConfig bool

	srcPath  string
	destPath string

	// configuration for the remote
	remoteParams []string

	// configuration for the encryption (if configured)
	encrypt       bool
	encryptParams []string
}

// Sync is running "rclone sync" to perform a synchronization of local files to remote destination
func (c SyncToRemoteCommand) Sync() error {
	runner := rclone.Runner{
		RenderConfig:  c.renderConfig,
		ConfigPath:    c.configPath,
		RemoteParams:  c.remoteParams,
		Encrypt:       c.encrypt,
		EncryptParams: c.encryptParams,
	}

	return runner.SyncToRemote(c.srcPath, c.destPath)
}
