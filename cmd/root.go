package cmd

import (
	"github.com/riotkit-org/volume-syncer/cmd/remote_to_local"
	syncToRemote "github.com/riotkit-org/volume-syncer/cmd/sync-to-remote"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Main() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume-syncer",
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

	return cmd
}
