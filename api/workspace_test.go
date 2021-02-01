package api

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cloud "github.com/mattermost/mattermost-cloud/model"

	"github.com/mattermost/pillar/mock"
	"github.com/mattermost/pillar/testlib"
	"github.com/mattermost/pillar/utils"
)

func TestWorkspace(t *testing.T) {
	logger := testlib.MakeLogger(t)

	ctrl := gomock.NewController(t)
	mockCloudClient := mock.NewMockCloudClient(ctrl)

	router := mux.NewRouter()
	Register(router, &Context{
		Logger:      logger,
		CloudClient: mockCloudClient,
	})
	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("list workspaces", func(t *testing.T) {
		client := NewClient(ts.URL)

		t.Run("success", func(t *testing.T) {
			mockInstallations := []*cloud.InstallationDTO{{Installation: &cloud.Installation{ID: "1234", DNS: "joram.cloud.mattermost.com"}}}
			mockCloudClient.EXPECT().GetInstallations(gomock.Any()).Times(1).Return(mockInstallations, nil)

			workspaces, err := client.ListWorkspaces(&cloud.GetInstallationsRequest{})
			assert.NoError(t, err)
			require.NotNil(t, workspaces)
			require.Len(t, workspaces, 1)
			assert.Equal(t, mockInstallations[0].ID, workspaces[0].ID)
		})

		t.Run("error getting installations", func(t *testing.T) {
			mockCloudClient.EXPECT().GetInstallations(gomock.Any()).Times(1).Return(nil, errors.New("some error"))
			workspaces, err := client.ListWorkspaces(&cloud.GetInstallationsRequest{})
			assert.Error(t, err)
			assert.Nil(t, workspaces)
		})
	})

	t.Run("get workspace", func(t *testing.T) {
		client := NewClient(ts.URL)

		t.Run("success", func(t *testing.T) {
			mockInstallation := &cloud.InstallationDTO{Installation: &cloud.Installation{ID: "installationid", DNS: "joram.cloud.mattermost.com", GroupID: utils.NewString("groupid")}}
			mockCloudClient.EXPECT().GetInstallation(gomock.Eq("installationid"), gomock.Any()).Times(1).Return(mockInstallation, nil)

			mockClusterInstallations := []*cloud.ClusterInstallation{{ID: "clusterinstallationid"}}
			mockCloudClient.EXPECT().GetClusterInstallations(gomock.Any()).Times(1).Return(mockClusterInstallations, nil)
			mockCloudClient.EXPECT().ExecClusterInstallationCLI(gomock.Eq("clusterinstallationid"), gomock.Eq("mmctl"), gomock.Any()).Times(1).Return([]byte("{\"ServiceSettings\":{}}"), nil)

			mockGroup := &cloud.Group{ID: "groupid"}
			mockCloudClient.EXPECT().GetGroup(gomock.Eq("groupid")).Times(1).Return(mockGroup, nil)

			workspace, err := client.GetWorkspace("installationid")
			assert.NoError(t, err)
			require.NotNil(t, workspace)
			assert.Equal(t, "installationid", workspace.ID)
			require.NotNil(t, workspace.Group)
			assert.Equal(t, "groupid", workspace.Group.ID)
			require.NotNil(t, workspace.Config)
			_, ok := workspace.Config["ServiceSettings"]
			assert.True(t, ok)
		})

		t.Run("not found", func(t *testing.T) {
			mockCloudClient.EXPECT().GetInstallation(gomock.Eq("installationid"), gomock.Any()).Times(1).Return(nil, nil)
			workspace, err := client.GetWorkspace("installationid")
			assert.Error(t, err)
			assert.Nil(t, workspace)
		})

		t.Run("no cluster installation", func(t *testing.T) {
			mockInstallation := &cloud.InstallationDTO{Installation: &cloud.Installation{ID: "installationid", DNS: "joram.cloud.mattermost.com", GroupID: utils.NewString("groupid")}}
			mockCloudClient.EXPECT().GetInstallation(gomock.Eq("installationid"), gomock.Any()).Times(1).Return(mockInstallation, nil)

			mockClusterInstallations := []*cloud.ClusterInstallation{}
			mockCloudClient.EXPECT().GetClusterInstallations(gomock.Any()).Times(1).Return(mockClusterInstallations, nil)

			mockGroup := &cloud.Group{ID: "groupid"}
			mockCloudClient.EXPECT().GetGroup(gomock.Eq("groupid")).Times(1).Return(mockGroup, nil)

			workspace, err := client.GetWorkspace("installationid")
			assert.Error(t, err)
			require.Nil(t, workspace)
		})

		t.Run("no group", func(t *testing.T) {
			mockInstallation := &cloud.InstallationDTO{Installation: &cloud.Installation{ID: "installationid", DNS: "joram.cloud.mattermost.com", GroupID: utils.NewString("groupid")}}
			mockCloudClient.EXPECT().GetInstallation(gomock.Eq("installationid"), gomock.Any()).Times(1).Return(mockInstallation, nil)

			mockClusterInstallations := []*cloud.ClusterInstallation{{ID: "clusterinstallationid"}}
			mockCloudClient.EXPECT().GetClusterInstallations(gomock.Any()).Times(1).Return(mockClusterInstallations, nil)
			mockCloudClient.EXPECT().ExecClusterInstallationCLI(gomock.Eq("clusterinstallationid"), gomock.Eq("mmctl"), gomock.Any()).Times(1).Return([]byte("{\"ServiceSettings\":{}}"), nil)

			mockCloudClient.EXPECT().GetGroup(gomock.Eq("groupid")).Times(1).Return(nil, nil)

			workspace, err := client.GetWorkspace("installationid")
			assert.NoError(t, err)
			require.NotNil(t, workspace)
			assert.Nil(t, workspace.Group)
		})

		t.Run("error getting installation", func(t *testing.T) {
			mockCloudClient.EXPECT().GetInstallation(gomock.Eq("installationid"), gomock.Any()).Times(1).Return(nil, errors.New("some error"))
			workspace, err := client.GetWorkspace("installationid")
			assert.Error(t, err)
			assert.Nil(t, workspace)
		})
	})
}
