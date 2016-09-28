package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	dockerclient "github.com/fsouza/go-dockerclient"
	dexec "github.com/ahmetalpbalkan/go-dexec"
	"sync"
)

var once sync.Once
var client *dockerclient.Client

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

func ConfigureDockerClientFromEnv() error {
	if client != nil {
		panic("Docker client has already been configured!")
	}

	var err error
	once.Do(func() {
		client, err = dockerclient.NewClientFromEnv()
	})

	return err
}

func DockerClientInfo() (string, error) {
	info, err := client.Info()
	prettyInfo, _ := json.MarshalIndent(info, "", "    ")
	return string(prettyInfo), err
}

// Pull pulls the image from the repository
func (dc *DockerConfig) PullOrRefresh() error {
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
	var images []dockerclient.APIImageSearch
	var err error
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
	var environment []string
	for k, v := range environmentVars {
		environment = append(environment, fmt.Sprintf("%s=%s", k, v))
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

	cmd := d.Command(m, command, arguments...)
	output, err := cmd.Output()
	if err != nil {
		return &RunOutput{Stdout: string(output[:]), Stderr: err.Error(), Error: true}, err
	}
	return &RunOutput{Stdout: string(output[:]), Error: false}, nil
}
