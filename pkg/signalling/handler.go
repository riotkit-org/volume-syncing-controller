package signalling

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/riotkit-org/volume-syncing-operator/pkg/cron"
	"github.com/riotkit-org/volume-syncing-operator/pkg/rclone"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// SetupInterruptSignal is setting up a graceful shutdown behavior
func SetupInterruptSignal(scheduler *cron.Scheduler, sync func() error, pidPath string) error {
	defer cleanUpOwnProcessId(pidPath)
	if err := storeOwnProcessId(pidPath); err != nil {
		return err
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		logrus.Info("Received INTERRUPT signal, disabling scheduler and performing last synchronization")

		scheduler.LockFromSchedulingNextIterations()
		syncErr := sync()
		if syncErr != nil {
			retriesAllowed := 30

			if syncErr == rclone.ErrAlreadyRunning {
				for {
					if retriesAllowed <= 0 {
						logrus.Error("Exceeded maximum retry count to synchronize on interruption")
						os.Exit(1)
					}

					time.Sleep(1)
					retriesAllowed -= 1
					syncErr = sync()
					if syncErr == nil {
						break
					}
					if syncErr != nil && syncErr != rclone.ErrAlreadyRunning {
						logrus.Errorf("Cannot execute synchronization on interrupt: %v", syncErr.Error())
						os.Exit(1)
					}
				}
			}
		}
		os.Exit(0)
	}()

	return nil
}

func storeOwnProcessId(pidPath string) error {
	pid := os.Getpid()
	if err := os.WriteFile(pidPath, []byte(fmt.Sprintf("%v", pid)), 0755); err != nil {
		return errors.Wrap(err, "Cannot write PID file")
	}
	return nil
}

func cleanUpOwnProcessId(pidPath string) {
	_ = os.Remove(pidPath)
}
