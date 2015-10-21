package remote_endpoint

import (
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"os"

	aperrors "gateway/errors"
)

type Script struct {
	Config struct {
		Interpreter string `json:"interpreter"`
		Timeout     int64  `json:"timeout"`
		FilePath    string `json:"filepath"`
		Script      string `json:"script"`
	} `json:"config"`
}

func (s *Script) Name() string {
	if s.Config.FilePath != "" {
		return s.Config.FilePath
	}

	sum := sha1.Sum([]byte(s.Config.Script))
	name := os.TempDir() + string(os.PathSeparator) + hex.EncodeToString(sum[:])
	if s.Config.Interpreter == "cmd.exe" {
		name += ".bat"
	}
	return name
}

func (s *Script) WriteFile() error {
	name := s.Name()
	if _, err := os.Stat(name); err == nil {
		return nil
	}

	if err := ioutil.WriteFile(name, []byte(s.Config.Script), 0777); err != nil {
		return err
	}

	return nil
}

func (s *Script) Inherit(parent *Script) {
	if s.Config.Interpreter == "" {
		s.Config.Interpreter = parent.Config.Interpreter
	}
	if s.Config.Timeout == 0 {
		s.Config.Timeout = parent.Config.Timeout
	}
	if s.Config.FilePath == "" {
		s.Config.FilePath = parent.Config.FilePath
	}
	if s.Config.Script == "" {
		s.Config.Script = parent.Config.Script
	}
}

func (s *Script) Validate(errors aperrors.Errors) {
	if s.Config.Interpreter == "" {
		errors.Add("interpreter", "must be set")
	}
	if s.Config.FilePath == "" && s.Config.Script == "" {
		errors.Add("filepath", "a File path or Script must be entered")
		errors.Add("script", "a File path or Script must be entered")
	}
	if s.Config.FilePath != "" {
		if _, err := os.Stat(s.Config.FilePath); err != nil {
			errors.Add("filepath", "must be a valid file")
		}
	}
}
