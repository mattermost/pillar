package api

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	cloud "github.com/mattermost/mattermost-cloud/model"
)

// Context provides the API with all necessary data and interfaces for responding to requests.
//
// It is cloned before each request, allowing per-request changes such as logger annotations.
type Context struct {
	RequestID   string
	Logger      logrus.FieldLogger
	CloudClient CloudClient
}

// Compile-time check to ensure CloudClient is implemented by cloud.Client
var _ CloudClient = &cloud.Client{}

// CloudClient is an interface that defines the client for connecting to the cloud provisioner.
type CloudClient interface {
	GetInstallation(string, *cloud.GetInstallationRequest) (*cloud.InstallationDTO, error)
	GetInstallations(*cloud.GetInstallationsRequest) ([]*cloud.InstallationDTO, error)
	GetClusterInstallations(*cloud.GetClusterInstallationsRequest) ([]*cloud.ClusterInstallation, error)
	ExecClusterInstallationCLI(string, string, []string) ([]byte, error)
	GetGroup(string) (*cloud.Group, error)
}

// Error represents an error response in the API
type Error struct {
	Message string
}

// Clone creates a shallow copy of context, allowing clones to apply per-request changes.
func (c *Context) Clone() *Context {
	return &Context{
		Logger:      c.Logger,
		CloudClient: c.CloudClient,
	}
}

func (c *Context) writeAndLogErrorWithFields(w http.ResponseWriter, err error, logFields logrus.Fields) {
	logger := c.Logger
	if logFields != nil {
		logger = logger.WithFields(logFields)
	}

	logger.Error(err)

	b, _ := json.Marshal(&Error{Message: err.Error()})
	w.Write(b)
}

func (c *Context) writeAndLogError(w http.ResponseWriter, err error) {
	c.writeAndLogErrorWithFields(w, err, nil)
}
