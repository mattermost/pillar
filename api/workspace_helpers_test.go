package api

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cloud "github.com/mattermost/mattermost-cloud/model"

	"github.com/mattermost/pillar/mock"
	"github.com/mattermost/pillar/utils"
)

func TestConvertInstallationToWorkspace(t *testing.T) {
	var installation *cloud.InstallationDTO

	t.Run("nil installation", func(t *testing.T) {
		workspace := convertInstallationToWorkspace(installation)
		assert.Nil(t, workspace)
	})

	t.Run("successful conversion", func(t *testing.T) {
		installation = &cloud.InstallationDTO{
			Installation: &cloud.Installation{
				ID:        "id",
				GroupID:   utils.NewString("groupid"),
				Version:   "version",
				DNS:       "joram.cloud.mattermost.com",
				Database:  cloud.InstallationDatabaseMultiTenantRDSPostgres,
				Filestore: cloud.InstallationFilestoreAwsS3,
				CreateAt:  1,
				DeleteAt:  0,
				Affinity:  cloud.InstallationAffinityMultiTenant,
			},
		}

		workspace := convertInstallationToWorkspace(installation)
		require.NotNil(t, workspace)
		assert.Equal(t, installation.ID, workspace.ID)
		assert.Equal(t, *installation.GroupID, workspace.GroupID)
		assert.Equal(t, installation.Version, workspace.Version)
		assert.Equal(t, installation.DNS, workspace.DNS)
		assert.Equal(t, installation.Database, workspace.Database)
		assert.Equal(t, installation.Filestore, workspace.Filestore)
		assert.Equal(t, installation.CreateAt, workspace.CreateAt)
		assert.Equal(t, installation.DeleteAt, workspace.DeleteAt)
		assert.Equal(t, WorkspaceEditionProfessional, workspace.Edition)

		installation.Affinity = cloud.InstallationAffinityIsolated

		workspace = convertInstallationToWorkspace(installation)
		require.NotNil(t, workspace)
		assert.Equal(t, WorkspaceEditionEnterprise, workspace.Edition)

		installation.GroupID = nil

		workspace = convertInstallationToWorkspace(installation)
		require.NotNil(t, workspace)
		assert.Empty(t, workspace.GroupID)
	})
}

func TestConvertInstallationsToWorkspaces(t *testing.T) {
	var installations []*cloud.InstallationDTO

	t.Run("nil installations", func(t *testing.T) {
		workspaces := convertInstallationsToWorkspaces(installations)
		assert.Nil(t, workspaces)
	})

	t.Run("successful conversion", func(t *testing.T) {
		installation1 := &cloud.InstallationDTO{Installation: &cloud.Installation{ID: "id1"}}
		installation2 := &cloud.InstallationDTO{Installation: &cloud.Installation{ID: "id2"}}
		installations = []*cloud.InstallationDTO{installation1, installation2}

		workspaces := convertInstallationsToWorkspaces(installations)
		require.NotNil(t, workspaces)
		require.Len(t, workspaces, len(installations))
		assert.Equal(t, installation1.ID, workspaces[0].ID)
		assert.Equal(t, installation2.ID, workspaces[1].ID)
	})
}

func TestConvertCloudGroupToGroup(t *testing.T) {
	var cloudGroup *cloud.Group

	t.Run("nil group", func(t *testing.T) {
		group := convertCloudGroupToGroup(cloudGroup)
		assert.Nil(t, group)
	})

	t.Run("succesful conversion", func(t *testing.T) {
		cloudGroup = &cloud.Group{
			ID:          "id",
			Name:        "name",
			Description: "description",
		}

		group := convertCloudGroupToGroup(cloudGroup)
		require.NotNil(t, group)
		assert.Equal(t, cloudGroup.ID, group.ID)
		assert.Equal(t, cloudGroup.Name, group.Name)
		assert.Equal(t, cloudGroup.Description, group.Description)
	})
}

func TestGetConfigForClusterInstallation(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCloudClient := mock.NewMockCloudClient(ctrl)

	t.Run("nil client", func(t *testing.T) {
		config, err := getConfigForClusterInstallation(nil, "junk")
		require.Error(t, err)
		assert.Equal(t, err.Error(), "CloudClient is nil")
		assert.Nil(t, config)
	})

	t.Run("success", func(t *testing.T) {
		mockCloudClient.
			EXPECT().
			ExecClusterInstallationCLI(gomock.Eq("id"), gomock.Eq("mmctl"), gomock.Any()).
			Times(1).
			Return([]byte("{\"ServiceSettings\":{}}"), nil)

		config, err := getConfigForClusterInstallation(mockCloudClient, "id")
		assert.NoError(t, err)
		require.NotNil(t, config)
		_, ok := config["ServiceSettings"]
		assert.True(t, ok)
	})

	t.Run("cloud client errors out", func(t *testing.T) {
		mockCloudClient.
			EXPECT().
			ExecClusterInstallationCLI(gomock.Eq("id"), gomock.Eq("mmctl"), gomock.Any()).
			Times(1).
			Return(nil, errors.New("some error"))

		config, err := getConfigForClusterInstallation(mockCloudClient, "id")
		require.Error(t, err)
		assert.Equal(t, err.Error(), "some error")
		assert.Nil(t, config)
	})
}
