package sync_to_remote

import (
	"github.com/riotkit-org/volume-syncing-controller/pkg/helpers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewSyncToRemoteCommand() *cobra.Command {
	var noTemplate bool
	var noDelete bool
	app := SyncToRemoteCommand{}

	command := &cobra.Command{
		Use:   "sync-to-remote",
		Short: "Copies files to remote filesystem",
		Run: func(command *cobra.Command, args []string) {
			app.renderConfig = !noTemplate
			app.cleanUp = !noDelete

			err := app.Sync()
			if err != nil {
				logrus.Fatal(err)
			}
		},
	}

	command.Flags().StringVarP(&app.configPath, "config-path", "c", helpers.GetEnvOrDefault("CONFIG_PATH", "rclone.conf").(string), "rclone configuration path (specify together with --no-template to use already prepared config)")
	command.Flags().BoolVarP(&noTemplate, "no-template", "", false, "Disables rendering of the rclone configuration file")
	command.Flags().StringVarP(&app.srcPath, "src", "s", "./", "Local path to copy files from")
	command.Flags().StringVarP(&app.destPath, "dst", "d", "/", "Target path")
	command.Flags().StringArrayVarP(&app.remoteParams, "param", "p", []string{}, "List of key=value settings for remote e.g. -p 'type=s3' -p 'provider=Minio' -p 'access_key_id=AKIAIOSFODNN7EXAMPLE'")
	command.Flags().StringArrayVarP(&app.encryptParams, "enc-param", "e", []string{}, "List of key=value settings for remote e.g. -p 'remote=remote:testbucket' -p 'password=xxxxxxxx'")
	command.Flags().StringVarP(&app.SchedulerExpression, "schedule", "", "", "Set to a valid crontab-like expression to schedule synchronization periodically")
	command.Flags().BoolVarP(&noDelete, "no-delete", "x", true, "Don't delete files in remote filesystem (may be dangerous if wrong path specified)")
	command.Flags().BoolVarP(&app.ForceSync, "force-even-if-remote-would-be-cleared", "f", true, "Force synchronize, even if it would mean to remove all files from remote")
	command.Flags().BoolVarP(&app.debug, "verbose", "v", true, "Increase verbosity")
	command.Flags().StringVarP(&app.pidPath, "pid-path", "", "/run/volume-syncing-controller.pid", "PID path")

	return command
}
