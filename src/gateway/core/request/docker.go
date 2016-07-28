package request

import (
	"encoding/json"
	"fmt"

	aperrors "gateway/errors"
	"gateway/model"

	"github.com/ahmetalpbalkan/dexec"
	docker "github.com/fsouza/go-dockerclient"
)

// DockerRequest is a request to a Docker container
type DockerRequest struct {
	Image     string   `json:"image"`
	Command   string   `json:"command"`
	Arguments []string `json:"arguments"`
	Advanced  bool     `json:"advanced"`
	Config    struct {
		Repository string `json:"repository,omitempty"`
		Tag        string `json:"tag,omitempty"`
		Username   string `json:"username,omitempty"`
		Password   string `json:"password,omitempty"`
	} `json:"config,omitempty"`
}

// DockerResponse is a response from a Docker container
type DockerResponse struct {
	Output []byte `json:"output"`
}

// JSON marshals a DockerResponse to JSON
func (dr *DockerResponse) JSON() ([]byte, error) {
	return json.Marshal(dr)
}

// Log formats the response's output
func (dr *DockerResponse) Log() string {
	return fmt.Sprintf("Output: %s", dr.Output)
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
	if other.Image != "" {
		dr.Image = other.Image
	}
	if other.Command != "" {
		dr.Command = other.Command
	}
	if areNotEqual(other.Arguments, dr.Arguments) {
		dr.Arguments = other.Arguments
	}
	if dr.Advanced != other.Advanced {
		dr.Advanced = other.Advanced
		if dr.Config.Repository != other.Config.Repository {
			dr.Config.Repository = other.Config.Repository
		}
		if dr.Config.Tag != other.Config.Tag {
			dr.Config.Tag = other.Config.Tag
		}
		if dr.Config.Username != other.Config.Username {
			dr.Config.Username = other.Config.Username
		}
		if dr.Config.Password != other.Config.Password {
			dr.Config.Password = other.Config.Password
		}
	}
}

// Log satisfies request.Request's Log method
func (dr *DockerRequest) Log(devMode bool) string {
	return fmt.Sprintf("Image: %s Command: %s Args: %v Advanced: %v Config: %+v", dr.Image, dr.Command, dr.Arguments, dr.Advanced, dr.Config)
}

// JSON satisfies request.Request's JSON method
func (dr *DockerRequest) JSON() ([]byte, error) {
	return json.Marshal(dr)
}

// Perform satisfies request.Request's Perform method
func (dr *DockerRequest) Perform() Response {
	if dr.Advanced {
		return performAdvanced(dr)
	}
	return performSimple(dr)
}

func performSimple(dr *DockerRequest) Response {
	client, _ := docker.NewClientFromEnv()
	d := dexec.Docker{client}
	m, _ := dexec.ByCreatingContainer(docker.CreateContainerOptions{Config: &docker.Config{Image: dr.Image}})

	cmd := d.Command(m, dr.Command, dr.Arguments...)
	output, err := cmd.Output()
	if err != nil {
		return NewErrorResponse(aperrors.NewWrapped("[docker] Error running command", err))
	}
	return &DockerResponse{Output: output}
}

func performAdvanced(dr *DockerRequest) Response {
	client, _ := docker.NewClientFromEnv()
	d := dexec.Docker{client}
	perr := client.PullImage(docker.PullImageOptions{
		Repository: dr.Config.Repository,
		Tag:        dr.Config.Tag,
	}, docker.AuthConfiguration{
		Username:      dr.Config.Username,
		Password:      dr.Config.Password,
		ServerAddress: dr.Config.Repository,
	})
	if perr != nil {
		return NewErrorResponse(aperrors.NewWrapped("[docker] Error pulling image", perr))
	}

	m, _ := dexec.ByCreatingContainer(docker.CreateContainerOptions{Config: &docker.Config{Image: dr.Image}})

	cmd := d.Command(m, dr.Command, dr.Arguments...)
	output, err := cmd.Output()
	if err != nil {
		return NewErrorResponse(aperrors.NewWrapped("[docker] Error running command", err))
	}
	return &DockerResponse{Output: output}
}

// areNotEqual checks the given slices for equality and returns true iff a and b are NOT equal
func areNotEqual(a, b []string) bool {
	return !areEqual(a, b)
}

// areEqual checks the given slices for equality and returns true iff a equals b
func areEqual(a, b []string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
