package docker

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ahmetalpbalkan/dexec"
	dockerclient "github.com/fsouza/go-dockerclient"
)

type DockerConfig struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	Registry   string `json:"registry,omitempty"`
	Memory     int64  `json:"-"`
	CPUShares  int64  `json:"-"`
}

type RunOutput struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Error  bool   `json:"error"`
}

// Pull pulls the image from the repository
func (dc *DockerConfig) PullOrRefresh() error {
	client, err := dockerclient.NewClientFromEnv()
	if err != nil {
		return errors.New("cannot initialize docker client.")
	}
	var perr error
	if dc.Tag == "latest" {
		perr = dc.PullImage(client)
	} else {
		_, ierr := client.InspectImage(dc.Image())
		if ierr == dockerclient.ErrNoSuchImage {
			perr = dc.PullImage(client)
		} else if ierr != nil {
			return ierr
		}
	}
	if perr != nil {
		return perr
	}
	return nil
}

// Image returns the image name of this docker request.
func (dc *DockerConfig) Image() string {
	var buffer bytes.Buffer
	buffer.WriteString(dc.Repository)
	buffer.WriteString(":")
	buffer.WriteString(dc.Tag)
	return buffer.String()
}

func (dc *DockerConfig) PullImage(client *dockerclient.Client) error {
	authConfig := dockerclient.AuthConfiguration{
		Username:      dc.Username,
		Password:      dc.Password,
		ServerAddress: dc.Registry,
	}
	perr := client.PullImage(dockerclient.PullImageOptions{
		Repository: dc.Repository,
		Tag:        dc.Tag,
		Registry:   dc.Registry,
	}, authConfig)

	if perr != nil {
		return perr
	}
	return nil
}

func (dc *DockerConfig) ImageExists() (bool, error) {
	client, err := dockerclient.NewClientFromEnv()
	if err != nil {
		return false, errors.New("cannot initialize docker client.")
	}
	var images []dockerclient.APIImageSearch
	if dc.Username != "" && dc.Password != "" || dc.Registry != "" {
		authConfig := dockerclient.AuthConfiguration{
			Username:      dc.Username,
			Password:      dc.Password,
			ServerAddress: dc.Registry,
		}
		images, err = client.SearchImagesEx(dc.Repository, authConfig)
	} else {
		images, err = client.SearchImages(dc.Repository)
	}
	if err != nil {
		return false, err
	}
	if images == nil || len(images) == 0 {
		return false, nil
	}
	return true, nil
}

func (dc *DockerConfig) Execute(command string, arguments []string, environmentVars map[string]string) (*RunOutput, error) {
	client, err := dockerclient.NewClientFromEnv()
	var environment []string
	for k, v := range environmentVars {
		environment = append(environment, fmt.Sprintf("%s=%s", k, v))
	}
	output := new(RunOutput)
	if err != nil {
		return output, err
	}

	d := dexec.Docker{Client: client}

	var perr error
	if dc.Tag == "latest" {
		perr = dc.PullImage(client)
	} else {
		_, ierr := client.InspectImage(dc.Image())
		if ierr == dockerclient.ErrNoSuchImage {
			perr = dc.PullImage(client)
		} else if ierr != nil {
			return nil, ierr
		}
	}
	if perr != nil {
		return nil, perr
	}
	m, err := dexec.ByCreatingContainer(dockerclient.CreateContainerOptions{
		Config: &dockerclient.Config{
			Image:     dc.Image(),
			Env:       environment,
			Memory:    dc.Memory * 1024 * 1024,
			CPUShares: dc.CPUShares,
		},
	})
	if err != nil {
		return nil, err
	}

	var stdout, stderr bytes.Buffer
	cmd := d.Command(m, command, arguments...)
	cmdErr := false
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		cmdErr = true
		if stderr.Len() < 1 {
			return &RunOutput{Stdout: stdout.String(), Stderr: err.Error(), Error: cmdErr}, err
		}
	}
	return &RunOutput{Stdout: stdout.String(), Stderr: stderr.String(), Error: cmdErr}, nil
}
