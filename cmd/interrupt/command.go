package interrupt

import (
	"github.com/pkg/errors"
	"os"
	"strconv"
	"syscall"
)

type Command struct {
	PidPath string
}

func (c *Command) Run() error {
	_, statErr := os.Stat(c.PidPath)
	if !os.IsExist(statErr) || statErr != nil {
		return errors.Wrapf(statErr, "Cannot find PID file at path '%s'. Is the process running?", c.PidPath)
	}

	pidByte, readErr := os.ReadFile(c.PidPath)
	if readErr != nil {
		return errors.Wrapf(readErr, "Cannot read PID file at path '%s'", c.PidPath)
	}

	pid, parseErr := strconv.Atoi(string(pidByte))
	if parseErr != nil {
		return errors.Wrapf(readErr, "Cannot parse PID file at path '%s' as integer", c.PidPath)
	}

	process, findErr := os.FindProcess(pid)
	if findErr != nil {
		return errors.Wrap(findErr, "Cannot find process. Maybe it exited just now?")
	}

	signalErr := process.Signal(syscall.SIGTERM)
	if signalErr != nil {
		return errors.Wrapf(signalErr, "Cannot kill process with pid '%v'", pid)
	}

	return nil
}
