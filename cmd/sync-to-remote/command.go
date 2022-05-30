package sync_to_remote

import (
	"github.com/riotkit-org/volume-syncer/pkg/cron"
	"github.com/riotkit-org/volume-syncer/pkg/rclone"
)

type SyncToRemoteCommand struct {
	configPath          string
	renderConfig        bool
	SchedulerExpression string

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
	if c.SchedulerExpression != "" {
		scheduler := cron.Scheduler{}
		return scheduler.SetupCron(c.SchedulerExpression, func() error {
			return c.sync()
		})
	}
	return c.sync()
}

func (c SyncToRemoteCommand) sync() error {
	runner := rclone.Runner{
		RenderConfig:  c.renderConfig,
		ConfigPath:    c.configPath,
		RemoteParams:  c.remoteParams,
		Encrypt:       c.encrypt,
		EncryptParams: c.encryptParams,
	}

	return runner.SyncToRemote(c.srcPath, c.destPath)
}
