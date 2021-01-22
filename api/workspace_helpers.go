package api

import (
	"encoding/json"

	cloud "github.com/mattermost/mattermost-cloud/model"
)

func convertInstallationToWorkspace(installation *cloud.InstallationDTO) *Workspace {
	edition := WorkspaceEditionProfessional
	if installation.Affinity == cloud.InstallationAffinityIsolated {
		edition = WorkspaceEditionEnterprise
	}

	groupID := ""
	if installation.GroupID != nil {
		groupID = *installation.GroupID
	}

	return &Workspace{
		ID:        installation.ID,
		GroupID:   groupID,
		Version:   installation.Version,
		DNS:       installation.DNS,
		Size:      installation.Size,
		Database:  installation.Database,
		Filestore: installation.Filestore,
		CreateAt:  installation.CreateAt,
		DeleteAt:  installation.DeleteAt,
		Edition:   edition,
	}
}

func convertInstallationsToWorkspaces(installations []*cloud.InstallationDTO) []*Workspace {
	workspaces := make([]*Workspace, len(installations))
	for index, installation := range installations {
		workspaces[index] = convertInstallationToWorkspace(installation)
	}
	return workspaces
}

func convertCloudGroupToGroup(cloudGroup *cloud.Group) *Group {
	return &Group{
		ID:          cloudGroup.ID,
		Name:        cloudGroup.Name,
		Description: cloudGroup.Description,
	}
}

func getConfigForClusterInstallation(client *cloud.Client, clusterInstallationID string) (map[string]interface{}, error) {
	cmdOutput, err := client.ExecClusterInstallationCLI(clusterInstallationID, "mmctl", []string{"config", "show", "--local"})
	if err != nil {
		return nil, err
	}

	configMap := map[string]interface{}{}
	err = json.Unmarshal(cmdOutput, &configMap)
	if err != nil {
		return nil, err
	}

	return configMap, nil
}
