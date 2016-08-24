package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"gateway/docker"
	aperrors "gateway/errors"
	"gateway/model"
)

// DockerRequest is a request to a Docker container
type DockerRequest struct {
	Repository  string   `json:"repository"`
	Tag         string   `json:"tag"`
	Command     string   `json:"command"`
	Arguments   []string `json:"arguments"`
	Environment []string `json:"environment"`
	Username    string   `json:"username,omitempty"`
	Password    string   `json:"password,omitempty"`
	Registry    string   `json:"registry,omitempty"`
}

// DockerResponse is a response from a Docker container
type DockerResponse struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Error  bool   `json:"error"`
}

// JSON marshals a DockerResponse to JSON
func (dr *DockerResponse) JSON() ([]byte, error) {
	return json.Marshal(dr)
}

// Log formats the response's output
func (dr *DockerResponse) Log() string {
	return fmt.Sprintf("Stdout: %s, Stderr: %s, Error: %t", dr.Stdout, dr.Stderr, dr.Error)
}

// NewDockerRequest creates a new Docker request
func NewDockerRequest(endpoint *model.RemoteEndpoint, data *json.RawMessage) (*DockerRequest, error) {
	request := new(DockerRequest)

	if err := json.Unmarshal(*data, request); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal request json: %v", err)
	}

	endpointData := new(DockerRequest)
	if err := json.Unmarshal(endpoint.Data, endpointData); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal endpoint data: %v", err)
	}
	request.updateWith(endpointData)

	if endpoint.SelectedEnvironmentData != nil {
		if err := json.Unmarshal(*endpoint.SelectedEnvironmentData, endpointData); err != nil {
			return nil, err
		}
		request.updateWith(endpointData)
	}

	return request, nil
}

func (dr *DockerRequest) updateWith(other *DockerRequest) {
	if other.Repository != "" {
		dr.Repository = other.Repository
	}
	if other.Tag != "" {
		dr.Tag = other.Tag
	}
	if other.Command != "" {
		dr.Command = other.Command
	}
	if len(other.Arguments) > 0 {
		dr.Arguments = other.Arguments
	}
	if len(other.Environment) > 0 {
		dr.Environment = other.Environment
	}
	if other.Username != "" && other.Password != "" {
		dr.Username = other.Username
		dr.Password = other.Password
	}
	if other.Registry != "" {
		dr.Registry = other.Registry
	}
}

// Log satisfies request.Request's Log method
func (dr *DockerRequest) Log(devMode bool) string {
	return fmt.Sprintf("Repository: %s Tag: %s Command: %s Args: %v", dr.Repository, dr.Tag, dr.Command, dr.Arguments)
}

// JSON satisfies request.Request's JSON method
func (dr *DockerRequest) JSON() ([]byte, error) {
	return json.Marshal(dr)
}

// Perform satisfies request.Request's Perform method
func (dr *DockerRequest) Perform() Response {
	if dr.Command == "" {
		return NewErrorResponse(errors.New("blank or nil commands are invalid"))
	}
	dc := &docker.DockerConfig{Repository: dr.Repository, Tag: dr.Tag, Username: dr.Username, Password: dr.Password, Registry: dr.Registry}
	runOutput, err := dc.Execute(dr.Command, dr.Arguments, dr.Environment)
	if err != nil {
		return NewErrorResponse(aperrors.NewWrapped("[docker] Error executing command in docker conatiner", err))
	}
	return &DockerResponse{Stdout: runOutput.Stdout, Stderr: runOutput.Stderr, Error: runOutput.Error}
}
