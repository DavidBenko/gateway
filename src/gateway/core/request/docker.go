package request

import (
	"bytes"
	"encoding/json"
	"fmt"

	aperrors "gateway/errors"
	"gateway/model"

	"github.com/ahmetalpbalkan/dexec"
	docker "github.com/fsouza/go-dockerclient"
)

// DockerRequest is a request to a Docker container
type DockerRequest struct {
	Endpoint  string `json:"endpoint"`
	Image     string `json:"image"`
	Command   string `json:"command"`
	UseTLS    bool   `json:"use_tls"`
	TLSConfig DockerTLS
}

// DockerTLS configuration for a TLS Docker request
type DockerTLS struct {
	CA          string `json:"ca"`
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"private_key"`
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
	if other.Endpoint != "" {
		dr.Endpoint = other.Endpoint
	}
	if other.Image != "" {
		dr.Image = other.Image
	}
	if other.Command != "" {
		dr.Command = other.Command
	}
	if other.UseTLS && !dr.UseTLS {
		dr.UseTLS = true
	}
	if other.TLSConfig.CA != "" {
		dr.TLSConfig.CA = other.TLSConfig.CA
	}
	if other.TLSConfig.Certificate != "" {
		dr.TLSConfig.Certificate = other.TLSConfig.Certificate
	}
	if other.TLSConfig.PrivateKey != "" {
		dr.TLSConfig.PrivateKey = other.TLSConfig.PrivateKey
	}
}

// Log satisfies request.Request's Log method
func (dr *DockerRequest) Log(devMode bool) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%s %s %s", dr.Command, dr.Endpoint, dr.Image))
	if devMode {
		buffer.WriteString("\nTLS CONFIG:\n")
		if dr.UseTLS {
			buffer.WriteString(fmt.Sprintf("CA:\n%s\nCERT:\n%s\nPK:\n%s\n", dr.TLSConfig.CA, dr.TLSConfig.Certificate, dr.TLSConfig.PrivateKey))
		} else {
			buffer.WriteString("UseTLS is set to false.")
		}
	}
	return buffer.String()
}

// JSON satisfies request.Request's JSON method
func (dr *DockerRequest) JSON() ([]byte, error) {
	return json.Marshal(dr)
}

// Perform satisfies request.Request's Perform method
func (dr *DockerRequest) Perform() Response {
	var cl *docker.Client
	if dr.UseTLS {
		cl, _ = docker.NewTLSClient(dr.Endpoint, dr.TLSConfig.Certificate, dr.TLSConfig.PrivateKey, dr.TLSConfig.CA)
	} else {
		cl, _ = docker.NewClient(dr.Endpoint)
	}

	d := dexec.Docker{cl}
	m, _ := dexec.ByCreatingContainer(docker.CreateContainerOptions{Config: &docker.Config{Image: dr.Image}})

	cmd := d.Command(m, dr.Command)
	output, err := cmd.Output()
	if err != nil {
		return NewErrorResponse(aperrors.NewWrapped("[docker] Error running command", err))
	}
	return &DockerResponse{Output: output}
}
