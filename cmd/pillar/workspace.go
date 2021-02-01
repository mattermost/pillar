package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cloud "github.com/mattermost/mattermost-cloud/model"

	"github.com/mattermost/pillar/api"
)

func init() {
	viper.SetEnvPrefix("PILLAR")
	viper.AutomaticEnv()

	workspaceCmd.PersistentFlags().String("server", defaultLocalServerAPI, "The pillar server whose API will be queried.")

	workspaceListCmd.Flags().String("owner", "", "The owner by which to filter workspaces.")
	workspaceListCmd.Flags().String("group", "", "The group ID by which to filter workspaces.")
	workspaceListCmd.Flags().Int("page", 0, "The page of workspaces to fetch, starting at 0.")
	workspaceListCmd.Flags().Int("per-page", 100, "The number of workspaces to fetch per page.")
	workspaceListCmd.Flags().Bool("include-deleted", false, "Whether to include deleted workspaces.")
	workspaceListCmd.Flags().String("dns", "", "The dns to filter results by.")
	workspaceCmd.AddCommand(workspaceListCmd)

	workspaceGetCmd.Flags().String("id", "", "ID of the workspace to get.")
	workspaceGetCmd.MarkFlagRequired("id")
	workspaceCmd.AddCommand(workspaceGetCmd)
}

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "View and edit workspaces.",
}

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workspaces.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		client := api.NewClient(serverAddress)

		owner, _ := command.Flags().GetString("owner")
		group, _ := command.Flags().GetString("group")
		page, _ := command.Flags().GetInt("page")
		perPage, _ := command.Flags().GetInt("per-page")
		includeDeleted, _ := command.Flags().GetBool("include-deleted")
		dns, _ := command.Flags().GetString("dns")
		workspaces, err := client.ListWorkspaces(&cloud.GetInstallationsRequest{
			OwnerID:                     owner,
			GroupID:                     group,
			IncludeGroupConfig:          true,
			IncludeGroupConfigOverrides: false,
			Page:                        page,
			PerPage:                     perPage,
			DNS:                         dns,
			IncludeDeleted:              includeDeleted,
		})
		if err != nil {
			return errors.Wrap(err, "failed to query workspaces")
		}

		err = printJSON(workspaces)
		if err != nil {
			return err
		}

		return nil
	},
}

var workspaceGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a workspace.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		client := api.NewClient(serverAddress)

		workspaceID, _ := command.Flags().GetString("id")
		workspace, err := client.GetWorkspace(workspaceID)
		if err != nil {
			return errors.Wrap(err, "failed to fetch workspace")
		}

		err = printJSON(workspace)
		if err != nil {
			return err
		}

		return nil
	},
}
