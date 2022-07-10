package cmd

import (
	"github.com/riotkit-org/volume-syncing-controller/cmd/interrupt"
	"github.com/riotkit-org/volume-syncing-controller/cmd/remote-to-local"
	"github.com/riotkit-org/volume-syncing-controller/cmd/serve"
	syncToRemote "github.com/riotkit-org/volume-syncing-controller/cmd/sync-to-remote"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Main() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume-syncing-controller",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				logrus.Errorf(err.Error())
			}
		},
	}
	cmd.AddCommand(syncToRemote.NewSyncToRemoteCommand())
	cmd.AddCommand(remote_to_local.NewRestoreCommand())
	cmd.AddCommand(serve.NewServeCommand())
	cmd.AddCommand(interrupt.NewInterruptionCommand())

	return cmd
}
