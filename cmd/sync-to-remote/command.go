package sync_to_remote

import (
	"github.com/pkg/errors"
	"github.com/riotkit-org/volume-syncing-operator/pkg/cron"
	"github.com/riotkit-org/volume-syncing-operator/pkg/helpers"
	"github.com/riotkit-org/volume-syncing-operator/pkg/rclone"
	"github.com/sirupsen/logrus"
)

type SyncToRemoteCommand struct {
	configPath          string
	renderConfig        bool
	SchedulerExpression string
	ForceSync           bool
	debug               bool

	srcPath  string
	destPath string

	// configuration for the remote
	remoteParams []string

	// configuration for the encryption (if configured)
	encrypt       bool
	encryptParams []string
}

// Sync is running "rclone sync" to perform a synchronization of local files to remote destination
func (c *SyncToRemoteCommand) Sync() error {
	if c.SchedulerExpression != "" {
		scheduler := cron.Scheduler{}
		return scheduler.SetupCron(c.SchedulerExpression, func() error {
			return c.sync()
		})
	}
	return c.sync()
}

func (c *SyncToRemoteCommand) sync() error {
	if err := c.validate(); err != nil {
		return errors.Wrap(err, "Error while trying to sync to remote")
	}

	runner := rclone.Runner{
		RenderConfig:     c.renderConfig,
		ConfigPath:       c.configPath,
		RemoteParams:     c.remoteParams,
		Encryption:       c.encrypt,
		EncryptionParams: c.encryptParams,
		Debug:            c.debug,
	}
	return runner.SyncToRemote(c.srcPath, c.destPath)
}

// validate will make sure that remote will not be accidentally deleted
func (c *SyncToRemoteCommand) validate() error {
	if c.ForceSync {
		logrus.Warn("--force-even-if-remote-would-be-cleared used, skipping validation, better know what you are doing")
		return nil
	}

	isEmpty, _ := helpers.IsLocalDirEmpty(c.srcPath)
	if isEmpty {
		return errors.New("Refusing to synchronize empty directory to remote - that would delete everything from remote filesystem")
	}
	return nil
}
