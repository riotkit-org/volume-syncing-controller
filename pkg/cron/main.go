package cron

import (
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type Scheduler struct {
	isRunning bool
	isLocked  bool
}

// SetupCron is setting up a scheduler that will run given action periodically
func (s *Scheduler) SetupCron(expression string, callback func() error) error {
	c := cron.New()
	_, err := c.AddFunc(expression, func() {
		logrus.Info("Trying to start task")

		if s.isLocked {
			logrus.Info("The scheduling was locked. Maybe the shutdown is in progress?")
			return
		}

		if s.isRunning {
			logrus.Warning("Existing job is already running, skipping next iteration")
			return
		}

		logrus.Info("Starting task")
		s.isRunning = true
		defer func() { s.isRunning = false }()

		if callbackErr := callback(); callbackErr != nil {
			logrus.WithError(callbackErr).Errorln("Cannot execute scheduled job")
		}
	})

	if err != nil {
		return errors.Wrap(err, "Cannot schedule job")
	}

	logrus.Infof("Scheduling task: %v", expression)
	c.Run()
	return nil
}

func (s *Scheduler) LockFromSchedulingNextIterations() {
	s.isLocked = true
}
