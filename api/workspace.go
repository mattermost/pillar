package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	cloud "github.com/mattermost/mattermost-cloud/model"
)

// initWorkspace registers workspace endpoints on the given router.
func initWorkspace(apiRouter *mux.Router, context *Context) {
	workspacesRouter := apiRouter.PathPrefix("/workspaces").Subrouter()
	workspacesRouter.Handle("/list", newAPIHandler(context, handleListWorkspaces)).Methods("POST")
	workspacesRouter.Handle("/{workspace}", newAPIHandler(context, handleGetWorkspace)).Methods("GET")
}

const (
	// WorkspaceEditionFree is the free cloud edition
	WorkspaceEditionFree = "Cloud Free"
	// WorkspaceEditionProfessional is the first paid cloud edition
	WorkspaceEditionProfessional = "Cloud Professional"
	// WorkspaceEditionEnterprise is the premium paid cloud edition
	WorkspaceEditionEnterprise = "Cloud Enterprise"
)

// Workspace is a respresentation of an installation without certain sensitive fields
// and data catered to be useful to the support team.
type Workspace struct {
	ID        string `json:"id"`
	GroupID   string `json:"group_id"`
	Version   string `json:"version"`
	DNS       string `json:"dns"`
	Size      string `json:"size"`
	Database  string `json:"database"`
	Filestore string `json:"filestore"`
	CreateAt  int64  `json:"create_at"`
	DeleteAt  int64  `json:"delete_at"`
	Edition   string `json:"edition"`
}

// handleListWorkspaces responds to POST /api/v1/workspaces/list, listing workspaces that match the filters.
func handleListWorkspaces(c *Context, w http.ResponseWriter, r *http.Request) {
	gi := &cloud.GetInstallationsRequest{}
	err := decodeJSON(gi, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, err)
		return
	}

	installations, err := c.CloudClient.GetInstallations(gi)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}

	resp := convertInstallationsToWorkspaces(installations)

	b, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// Group is a respresentation of an installation group without certain sensitive fields
// and data catered to be useful to the support team.
type Group struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// WorkspaceDetailed contains a workspace and extra detailed and related data for it.
type WorkspaceDetailed struct {
	*Workspace
	Group  *Group                 `json:"group"`
	Config map[string]interface{} `json:"config"`
}

// handleGetWorkspace responds to GET /api/v1/workspaces/{id}, getting a workspace and a bunch of contextual data for it.
func handleGetWorkspace(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspace"]
	c.Logger = c.Logger.WithField("workspace", workspaceID)

	installation, err := c.CloudClient.GetInstallation(workspaceID, &cloud.GetInstallationRequest{IncludeGroupConfig: true})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}
	if installation == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	workspace := convertInstallationToWorkspace(installation)

	configChan := make(chan error, 1)
	groupChan := make(chan error, 1)

	var config map[string]interface{}
	go func() {
		clusterInstallations, err := c.CloudClient.GetClusterInstallations(&cloud.GetClusterInstallationsRequest{InstallationID: workspace.ID, PerPage: 1000})
		if err != nil {
			configChan <- err
			return
		}

		if len(clusterInstallations) == 0 {
			configChan <- errors.New("workspace does not have a cluster installation")
			return
		}

		clusterInstallation := clusterInstallations[0]
		config, err = getConfigForClusterInstallation(c.CloudClient, clusterInstallation.ID)
		if err != nil {
			configChan <- err
			return
		}

		configChan <- nil
	}()

	var group *Group
	go func() {
		cloudGroup, err := c.CloudClient.GetGroup(workspace.GroupID)
		if err != nil {
			groupChan <- err
			return
		}

		group = convertCloudGroupToGroup(cloudGroup)
		groupChan <- nil
	}()

	configErr := <-configChan
	if configErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, configErr)
		return
	}

	groupErr := <-groupChan
	if groupErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, groupErr)
		return
	}

	workspaceDetailed := &WorkspaceDetailed{
		Workspace: workspace,
		Group:     group,
		Config:    config,
	}

	b, err := json.Marshal(workspaceDetailed)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
