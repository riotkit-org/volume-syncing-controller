package rclone

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

type Runner struct {
	RenderConfig bool
	ConfigPath   string

	// configuration for the remote
	RemoteParams []string

	// configuration for the encryption (if configured)
	Encryption       bool
	EncryptionParams []string

	Debug bool
}

// SyncToRemote invokes a "rclone sync" to remote destination
func (r *Runner) SyncToRemote(localPath string, targetPath string) error {
	return r.performFilesCopying("sync", localPath, r.buildRemotePath(targetPath))
}

// CopyToRemote invokes a "rclone copy" to remote destination
func (r *Runner) CopyToRemote(localPath string, targetPath string) error {
	return r.performFilesCopying("copy", localPath, r.buildRemotePath(targetPath))
}

// SyncFromRemote invokes a "rclone sync" to bring files back from remote storage
func (r *Runner) SyncFromRemote(remotePath string, localTargetPath string) error {
	return r.performFilesCopying("sync", r.buildRemotePath(remotePath), localTargetPath)
}

// CopyFromRemote invokes a "rclone copy" to bring files back from remote storage
func (r *Runner) CopyFromRemote(remotePath string, localTargetPath string) error {
	return r.performFilesCopying("copy", r.buildRemotePath(remotePath), localTargetPath)
}

func (r *Runner) buildRemotePath(remotePath string) string {
	// encrypted remote wraps a remote that already points to a target directory
	// so there in result we use root directory as we are already in a "chroot"
	if r.Encryption {
		return r.getRemoteName() + ":/"
	}

	return r.getRemoteName() + ":/" + strings.TrimLeft(remotePath, "/")
}

// performFilesCopying invokes a "rclone" command
func (r *Runner) performFilesCopying(action string, from string, to string) error {
	logrus.Infof("Performing %s from '%s' to '%s'", action, from, to)

	if r.RenderConfig {
		configErr := r.createConfig()
		defer r.cleanUpConfig()

		if configErr != nil {
			return errors.Wrap(configErr, "Cannot sync/copy")
		}
	}

	params := []string{action, "--create-empty-src-dirs", from, to}
	if r.Debug {
		params = append(params, "-vv")
	}
	return r.rclone(params...)
}

// IsDirEmpty checks if remote directory is empty
func (r *Runner) IsDirEmpty(path string) (bool, error) {
	if r.RenderConfig {
		configErr := r.createConfig()
		defer r.cleanUpConfig()

		if configErr != nil {
			return true, errors.Wrap(configErr, "Cannot sync/copy")
		}
	}

	proc := exec.Command("rclone", "ls", r.getRemoteName()+":/"+strings.TrimLeft(path, "/"))
	proc.Stderr = os.Stderr
	proc.Stdin = os.Stdin
	proc.Env = append(os.Environ(), "RCLONE_CONFIG="+r.ConfigPath)

	out, err := proc.Output()
	if err != nil {
		return true, errors.Wrap(err, "Cannot run rclone")
	}

	return strings.Trim(string(out), " \n") == "", nil
}

func (r *Runner) rclone(args ...string) error {
	logrus.Debug("rclone", args)

	proc := exec.Command("rclone", args...)
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	proc.Stdin = os.Stdin
	proc.Env = append(os.Environ(), "RCLONE_CONFIG="+r.ConfigPath)

	if err := proc.Run(); err != nil {
		return errors.Wrap(err, "Cannot run rclone")
	}
	return nil
}

// getRemoteName decides if we are using encryption
func (r *Runner) getRemoteName() string {
	if r.Encryption {
		return "remote_encrypted"
	}
	return "remote"
}

// createConfig is writing a configuration file required by "rclone" process
func (r *Runner) createConfig() error {
	logrus.Debug("Rendering config file...")

	ini := "[remote]\n"

	for _, param := range r.RemoteParams {
		parts := strings.Split(param, "=")
		if len(parts) < 2 {
			return errors.Errorf("Missing value in '%v'", param)
		}

		ini += param + "\n"
	}

	if r.Encryption {
		ini += "[remote_encrypted]\ntype=crypt\n"

		for _, param := range r.EncryptionParams {
			parts := strings.Split(param, "=")
			if len(parts) < 2 {
				return errors.Errorf("Missing value in '%v'", param)
			}

			ini += param + "\n"
		}
	}

	if err := os.WriteFile(r.ConfigPath, []byte(ini), 0755); err != nil {
		return errors.Wrap(err, "Cannot write configuration file")
	}

	return nil
}

// cleanUpConfig is deleting configuration file (it contains sensitive information such as credentials)
func (r *Runner) cleanUpConfig() {
	if err := os.Remove(r.ConfigPath); err != nil {
		logrus.Warnf("Cannot clean up configuration file: %v", err.Error())
	}
}
