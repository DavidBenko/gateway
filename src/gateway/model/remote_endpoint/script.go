package remote_endpoint

import (
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"os"
	"runtime"

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

var interpreters map[string][]string = map[string][]string{
	"linux":   []string{"sh", "bash"},
	"windows": []string{"cmd.exe"},
	"darwin":  []string{"sh"},
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

func (s *Script) UpdateWith(child *Script) {
	if child.Config.Interpreter != "" {
		s.Config.Interpreter = child.Config.Interpreter
	}
	if child.Config.Timeout != 0 {
		s.Config.Timeout = child.Config.Timeout
	}
	if child.Config.FilePath != "" {
		s.Config.FilePath = child.Config.FilePath
	}
	if child.Config.Script != "" {
		s.Config.Script = child.Config.Script
	}
}

func (s *Script) Validate() aperrors.Errors {
	errors := make(aperrors.Errors)
	if s.Config.Interpreter == "" {
		errors.Add("interpreter", "must be set")
	}

	if interps, ok := interpreters[runtime.GOOS]; !ok {
		errors.Add("interpreter", "unable to determine supported interpreters")
	} else {
		found := false
		for _, interp := range interps {
			if interp == s.Config.Interpreter {
				found = true
			}
		}
		if !found {
			errors.Add("interpreter", "not supported on this platform")
		}
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
	if !errors.Empty() {
		return errors
	}
	return nil
}
