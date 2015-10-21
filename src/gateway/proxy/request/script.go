package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"gateway/model"
	"gateway/model/remote_endpoint"
)

type ScriptRequest struct {
	remote_endpoint.Script
	Env map[string]string `json:"env"`
}

type ScriptResponse struct {
	Stdout string `json:"stdout"`
}

func (s *ScriptRequest) Perform() Response {
	response := &ScriptResponse{}
	var cmd *exec.Cmd
	if s.Config.Interpreter == "cmd.exe" {
		cmd = exec.Command("cmd", "/C", s.Name())
	} else {
		cmd = exec.Command(s.Config.Interpreter, s.Name())
	}
	cmd.Env = make([]string, len(s.Env))
	c := 0
	for key, value := range s.Env {
		cmd.Env[c] = fmt.Sprintf("%v=%v", key, value)
		c++
	}
	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout

	err := cmd.Start()
	if err != nil {
		return NewErrorResponse(err)
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	timeout := time.Duration(s.Config.Timeout)
	if timeout == 0 {
		timeout = 60
	}
	ticker := time.NewTicker(timeout * time.Second)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		if err := cmd.Process.Kill(); err != nil {
			return NewErrorResponse(err)
		}
		return NewErrorResponse(fmt.Errorf("the script timed out"))
	case err := <-done:
		if err != nil {
			return NewErrorResponse(err)
		}
	}

	response.Stdout = stdout.String()
	return response
}

func (h *ScriptRequest) Log(devMode bool) string {
	return ""
}

func (s *ScriptResponse) JSON() ([]byte, error) {
	return json.Marshal(&s)
}

func (r *ScriptResponse) Log() string {
	return ""
}

func NewScriptRequest(endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	request := &ScriptRequest{}
	if err := json.Unmarshal(*data, request); err != nil {
		return nil, err
	}

	endpointData := &ScriptRequest{}
	if err := json.Unmarshal(endpoint.Data, endpointData); err != nil {
		return nil, err
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

func (s *ScriptRequest) updateWith(endpointData *ScriptRequest) {
	if endpointData.Config.Interpreter != "" {
		s.Config.Interpreter = endpointData.Config.Interpreter
	}
	if endpointData.Config.Timeout != 0 {
		s.Config.Timeout = endpointData.Config.Timeout
	}
	if endpointData.Config.FilePath != "" {
		s.Config.FilePath = endpointData.Config.FilePath
	}
	if endpointData.Config.Script != "" {
		s.Config.Script = endpointData.Config.Script
	}
}
