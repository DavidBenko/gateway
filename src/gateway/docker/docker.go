package docker

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"gateway/config"
	"gateway/logreport"

	dockerclient "github.com/fsouza/go-dockerclient"
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
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	Logs       string `json:"logs"`
	StatusCode int    `json:"status_code"`
	Error      bool   `json:"error"`
}

func (r *RunOutput) Parts() ([]string, string) {
	parts := strings.Split(r.Stdout, "\x00\x00\x00\x00\x00\x00\x00\x00")
	lines := strings.Split(parts[0], "\n")
	for _, line := range strings.Split(r.Stderr, "\n") {
		if len(line) > 0 {
			lines = append(lines, "[error] "+line)
		}
	}
	if len(parts) > 1 {
		return lines, parts[1]
	}
	return lines, "null"
}

func ConfigureDockerClient(dockerConfig config.Docker) error {
	if client != nil {
		panic("Docker client has already been configured!")
	}

	var err error
	once.Do(func() {
		if dockerConfig.Host == "" {
			client, err = dockerclient.NewClientFromEnv()
		} else {
			if dockerConfig.Tls {
				if dockerConfig.TlsCertFile != "" {
					if dockerConfig.TlsKeyFile == "" || dockerConfig.TlsCaCertFile == "" {
						err = errors.New("Both a key file and ca cert file are required when Docker TLS is configured and a cert file is provided")
						return
					}
					client, err = dockerclient.NewTLSClient(dockerConfig.Host, dockerConfig.TlsCertFile, dockerConfig.TlsKeyFile, dockerConfig.TlsCaCertFile)
				} else if dockerConfig.TlsCertContent != "" {
					if dockerConfig.TlsKeyContent == "" || dockerConfig.TlsCaCertContent == "" {
						err = errors.New("Both key file content and ca cert file content are required when Docker TLS is configured and cert file content is provided")
						return
					}
					var certContents, caCertContents, keyContents []byte
					certContents, err = base64.StdEncoding.DecodeString(dockerConfig.TlsCertContent)
					caCertContents, err = base64.StdEncoding.DecodeString(dockerConfig.TlsCaCertContent)
					keyContents, err = base64.StdEncoding.DecodeString(dockerConfig.TlsKeyContent)
					client, err = dockerclient.NewTLSClientFromBytes(dockerConfig.Host, certContents, keyContents, caCertContents)
				} else {
					err = errors.New("Docker TLS is configured but no cert file or cert file content is provided")
					return
				}
			} else {
				client, err = dockerclient.NewClient(dockerConfig.Host)
			}
		}
	})

	return err
}

func DockerClientInfo() (string, error) {
	info, err := client.Info()
	prettyInfo, _ := json.MarshalIndent(info, "", "    ")
	return string(prettyInfo), err
}

func Available() bool {
	return client != nil
}

func BuildImage(options dockerclient.BuildImageOptions) error {
	return client.BuildImage(options)
}

func InspectImage(name string) (*dockerclient.Image, error) {
	return client.InspectImage(name)
}

func ExecuteImage(name string, memory, cpuShares, timeout int64, input interface{}) (*RunOutput, error) {
	var stdout, stderr, containerLogs bytes.Buffer

	container, err := client.CreateContainer(dockerclient.CreateContainerOptions{
		Config: &dockerclient.Config{
			Image:        name,
			StdinOnce:    true,
			OpenStdin:    true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Memory:       memory * 1024 * 1024,
			CPUShares:    cpuShares,
		},
	})

	if err != nil {
		return nil, err
	}

	defer func() {
		if er := client.RemoveContainer(dockerclient.RemoveContainerOptions{
			ID: container.ID,
		}); er != nil {
			logreport.Printf("%s Could not remove container %s: %s", config.System, container.ID, err.Error())
		}
	}()

	if err = client.StartContainer(container.ID, &dockerclient.HostConfig{}); err != nil {
		return nil, err
	}

	data, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var stdin = strings.NewReader(string(data))

	go func() {
		if er := client.AttachToContainer(dockerclient.AttachToContainerOptions{
			Container:    container.ID,
			Stream:       true,
			Stdin:        true,
			Stdout:       true,
			Stderr:       true,
			InputStream:  stdin,
			OutputStream: &stdout,
			ErrorStream:  &stderr,
		}); er != nil {
			logreport.Printf("%s Could not attach to container %s: %s", config.System, container.ID, err.Error())
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	code, err := client.WaitContainerWithContext(container.ID, ctx)
	if err != nil {
		client.StopContainer(container.ID, 0)
		return nil, err
	}

	err = client.Logs(dockerclient.LogsOptions{
		Container:    container.ID,
		Stdout:       true,
		Stderr:       true,
		OutputStream: &containerLogs,
		ErrorStream:  &containerLogs,
	})

	if err != nil {
		return nil, err
	}

	dockerErr := false
	if stderr.Len() > 0 {
		dockerErr = true
	}

	return &RunOutput{Stdout: stdout.String(), Stderr: stderr.String(), Logs: containerLogs.String(), StatusCode: code, Error: dockerErr}, nil
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

	var stdout, stderr, containerLogs bytes.Buffer

	container, err := client.CreateContainer(dockerclient.CreateContainerOptions{
		Config: &dockerclient.Config{
			Image:        dc.Image(),
			StdinOnce:    true,
			OpenStdin:    true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Env:          environment,
			Memory:       dc.Memory * 1024 * 1024,
			CPUShares:    dc.CPUShares,
		},
	})

	if err != nil {
		return nil, err
	}

	defer func() {
		if err = client.RemoveContainer(dockerclient.RemoveContainerOptions{
			ID: container.ID,
		}); err != nil {
			logreport.Printf("%s Could not remove container %s: %s", config.System, container.ID, err.Error())
		}
	}()

	if err = client.StartContainer(container.ID, &dockerclient.HostConfig{}); err != nil {
		return nil, err
	}

	var stdin = strings.NewReader(strings.Join(append([]string{command}, arguments...), " "))

	go func() {
		if err = client.AttachToContainer(dockerclient.AttachToContainerOptions{
			Container:    container.ID,
			Stream:       true,
			Stdin:        true,
			Stdout:       true,
			Stderr:       true,
			InputStream:  stdin,
			OutputStream: &stdout,
			ErrorStream:  &stderr,
		}); err != nil {
			logreport.Printf("%s Could not attach to container %s: %s", config.System, container.ID, err.Error())
		}
	}()

	code, err := client.WaitContainer(container.ID)
	if err != nil {
		return nil, err
	}

	err = client.Logs(dockerclient.LogsOptions{
		Container:    container.ID,
		Stdout:       true,
		Stderr:       true,
		OutputStream: &containerLogs,
		ErrorStream:  &containerLogs,
	})

	if err != nil {
		return nil, err
	}

	dockerErr := false
	if stderr.Len() > 0 {
		dockerErr = true
	}

	return &RunOutput{Stdout: stdout.String(), Stderr: stderr.String(), Logs: containerLogs.String(), StatusCode: code, Error: dockerErr}, nil
}
