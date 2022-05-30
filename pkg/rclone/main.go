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
	Encrypt       bool
	EncryptParams []string
}

// SyncToRemote invokes a "rclone sync" to remote destination
func (r *Runner) SyncToRemote(localPath string, targetPath string) error {
	return r.sync(localPath, r.getRemoteName()+":/"+strings.TrimLeft(targetPath, "/"))
}

// sync invokes a "rclone sync" command
func (r *Runner) sync(from string, to string) error {
	if r.RenderConfig {
		configErr := r.createConfig()
		defer r.cleanUpConfig()

		if configErr != nil {
			return errors.Wrap(configErr, "Cannot sync")
		}
	}

	proc := exec.Command("rclone", "sync", "-vv", "--create-empty-src-dirs", from, to)
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	proc.Stdin = os.Stdin
	proc.Env = append(os.Environ(), "RCLONE_CONFIG="+r.ConfigPath)

	if err := proc.Run(); err != nil {
		return errors.Wrap(err, "Cannot run 'rclone sync'")
	}
	return nil
}

// getRemoteName decides if we are using encryption
func (r *Runner) getRemoteName() string {
	if r.Encrypt {
		return "remote_encrypted"
	}
	return "remote"
}

// createConfig is writing a configuration file required by "rclone" process
func (r *Runner) createConfig() error {
	ini := "[remote]\n"

	for _, param := range r.RemoteParams {
		parts := strings.Split(param, "=")
		if len(parts) < 2 {
			return errors.Errorf("Missing value in '%v'", param)
		}

		ini += param + "\n"
	}

	if r.Encrypt {
		ini += "[remote_encrypted]\ntype=crypt\n"

		for _, param := range r.EncryptParams {
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
