package serve

import (
	"github.com/riotkit-org/volume-syncing-controller/pkg/helpers"
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
	command.Flags().StringVarP(&app.LogLevel, "log-level", "l", helpers.GetEnvOrDefault("LOG_LEVEL", "info").(string), "Logging level: error, warn, info, debug")
	command.Flags().BoolVarP(&app.LogJSON, "log-json", "", helpers.GetEnvOrDefault("LOG_JSON", false).(bool), "Log in JSON format")
	command.Flags().BoolVarP(&app.TLS, "tls", "t", helpers.GetEnvOrDefault("USE_TLS", false).(bool), "Use TLS (requires certificates)")
	command.Flags().StringVarP(&app.TLSCrtPath, "tls-crt", "", helpers.GetEnvOrDefault("TLS_CRT_PATH", "/etc/admission-webhook/tls/tls.crt").(string), "tls.crt")
	command.Flags().StringVarP(&app.TLSKeyPath, "tls-key", "", helpers.GetEnvOrDefault("TLS_KEY_PATH", "/etc/admission-webhook/tls/tls.key").(string), "tls.key")
	command.Flags().StringVarP(&app.Image, "image", "i", helpers.GetEnvOrDefault("IMAGE", "ghcr.io/riotkit-org/volume-syncing-controller:snapshot").(string), "Docker image")

	return command
}
