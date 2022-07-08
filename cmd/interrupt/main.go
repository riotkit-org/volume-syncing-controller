package interrupt

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewInterruptionCommand() *cobra.Command {
	app := Command{}

	command := &cobra.Command{
		Use:   "interrupt",
		Short: "Send interruption signal to the running process of synchronization. Used by Kubernetes to notify about a case, when the Pod is going down",
		Run: func(command *cobra.Command, args []string) {
			err := app.Run()
			if err != nil {
				logrus.Fatal(err)
			}
		},
	}
	command.Flags().StringVarP(&app.PidPath, "pid-path", "p", "/run/volume-syncing-operator.pid", "PID path")

	return command
}
