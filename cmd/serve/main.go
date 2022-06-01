package serve

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewServeCommand() *cobra.Command {
	app := Command{}

	command := &cobra.Command{
		Use:   "serve",
		Short: "Serve HTTP handler for the Kubernetes Admission Webhook",
		Run: func(command *cobra.Command, args []string) {
			err := app.Run()
			if err != nil {
				logrus.Fatal(err)
			}
		},
	}

	// command.Flags().StringVarP(&app.configPath, "config-path", "c", helpers.GetEnvOrDefault("CONFIG_PATH", "rclone.conf").(string), "rclone configuration path (specify together with --no-template to use already prepared config)")

	return command
}
