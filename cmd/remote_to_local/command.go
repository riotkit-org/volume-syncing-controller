package remote_to_local

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/riotkit-org/volume-syncing-operator/pkg/cron"
	"github.com/riotkit-org/volume-syncing-operator/pkg/helpers"
	"github.com/riotkit-org/volume-syncing-operator/pkg/rclone"
	"github.com/riotkit-org/volume-syncing-operator/pkg/signalling"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"syscall"
)

type Command struct {
	configPath          string
	renderConfig        bool
	SchedulerExpression string
	cleanUp             bool
	forceCleanUp        bool
	debug               bool
	pidPath             string

	srcPath       string
	destLocalPath string

	// configuration for the remote
	remoteParams []string

	// configuration for the encryption (if configured)
	encryptionParams []string
}

// Sync is running "rclone sync" to perform restore from a remote storage
func (c *Command) Sync() error {
	if c.SchedulerExpression != "" {
		scheduler := cron.Scheduler{}
		if err := signalling.SetupInterruptSignal(&scheduler, func() error { return c.sync() }, c.pidPath); err != nil {
			return err
		}

		return scheduler.SetupCron(c.SchedulerExpression, func() error {
			return c.sync()
		})
	}
	return c.sync()
}

func (c *Command) sync() error {
	runner := rclone.Runner{
		RenderConfig:     c.renderConfig,
		ConfigPath:       c.configPath,
		RemoteParams:     c.remoteParams,
		EncryptionParams: c.encryptionParams,
		Debug:            c.debug,
	}

	if len(c.encryptionParams) > 0 {
		runner.Encryption = true
	}
	if c.cleanUp {
		if err := c.validateBeforeDelete(runner); err != nil {
			return errors.Wrap(err, "Pre-delete validation failed")
		}

		return runner.SyncFromRemote(c.srcPath, c.destLocalPath)
	}

	return runner.CopyFromRemote(c.srcPath, c.destLocalPath)
}

// validateBeforeDelete tries to avoid disaster resulting from careless usage
func (c *Command) validateBeforeDelete(runner rclone.Runner) error {
	if c.forceCleanUp {
		logrus.Warn("Use --force-delete-local-dir with caution! You can accidentally wipe your drive.")
		return nil
	}

	//
	// Is remote dir empty? If yes, then it would make local dir empty as well...
	// When LOCAL DIR is NOT EMPTY, then do not allow emptying it
	//
	isLocalDirEmpty, _ := helpers.IsLocalDirEmpty(c.srcPath)
	if !isLocalDirEmpty {
		isRemoteDirEmpty, checkErr := runner.IsDirEmpty(c.srcPath)
		if checkErr != nil {
			return errors.Wrap(checkErr, "Pre-delete validation failed")
		}
		if isRemoteDirEmpty {
			return errors.New("Pre-delete validation failed, remote directory is empty - that would delete your local all files")
		}
	}

	//
	// Do we want to synchronize root of the filesystem or /usr/bin or other critical directory?
	//
	p := strings.ToLower(strings.Trim(c.destLocalPath, "/"))
	if p == "" || p == "root" || p == "usr/bin" || p == "bin" || p == "usr/lib" || p == "home" {
		return errors.New("Are you sure you want to delete this local source directory before synchronization? Use --force-delete-local-dir to force override")
	}

	//
	// Permissions validation
	//
	stat, _ := os.Stat(c.destLocalPath)
	sys := stat.Sys().(*syscall.Stat_t)
	if sys.Uid == 0 {
		return errors.New("Are you sure you want to delete directory owned by root? Use --force-delete-local-dir to force override")
	}
	if fmt.Sprintf("%v", sys.Uid) != fmt.Sprintf("%v", os.Getuid()) {
		return errors.New(fmt.Sprintf("Are you sure you want to delete directory owned by other user? (uid=%v vs uid=%v) Use --force-delete-local-dir to force override", sys.Uid, os.Getuid()))
	}

	return nil
}
