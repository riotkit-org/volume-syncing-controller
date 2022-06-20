package remote_to_local

import (
	"github.com/riotkit-org/volume-syncing-operator/pkg/helpers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewRestoreCommand() *cobra.Command {
	var noTemplate bool
	var noDelete bool
	app := Command{}

	command := &cobra.Command{
		Use:   "remote-to-local-sync",
		Short: "Restore files from remote storage into the local filesystem",
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
	command.Flags().StringVarP(&app.srcPath, "src", "s", "./", "Source path on remote storage")
	command.Flags().StringVarP(&app.destLocalPath, "dst", "d", "/", "Local target path")
	command.Flags().StringSliceVarP(&app.remoteParams, "param", "p", []string{}, "List of key=value settings for remote e.g. -p 'type=s3' -p 'provider=Minio' -p 'access_key_id=AKIAIOSFODNN7EXAMPLE'")
	command.Flags().StringSliceVarP(&app.encryptionParams, "enc-param", "e", []string{}, "List of key=value settings for remote e.g. -p 'remote=remote:testbucket' -p 'password=xxxxxxxx'")
	command.Flags().StringVarP(&app.SchedulerExpression, "schedule", "", "", "Set to a valid crontab-like expression to schedule synchronization periodically")
	command.Flags().BoolVarP(&noDelete, "no-delete", "x", false, "Don't delete files in local filesystem (may be dangerous if wrong path specified)")
	command.Flags().BoolVarP(&app.forceCleanUp, "force-delete-local-dir", "n", false, "Force delete local files that are not present on remote")
	command.Flags().BoolVarP(&app.debug, "verbose", "v", false, "Increase verbosity")
	// todo: --fsnotify

	return command
}
