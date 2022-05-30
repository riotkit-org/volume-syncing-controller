package cmd

import (
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

	return cmd
}
