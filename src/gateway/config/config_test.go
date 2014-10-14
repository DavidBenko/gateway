package config

import (
	"strings"
	"testing"
)

func TestFindConfigFile(t *testing.T) {
	cases := map[string]string{
		"":                  "",
		"-port 4000":        "",
		"-config":           "",
		"--config":          "",
		"-config foo":       "foo",
		"-config=foo":       "foo",
		"--config foo":      "foo",
		"--config=foo":      "foo",
		"-config=foo -test": "foo",
		"-test -config=foo": "foo",
	}

	for line, expectedFileName := range cases {
		filename := findConfigFile(strings.Split(line, " "))
		if filename != expectedFileName {
			t.Errorf("findConfigFile returned '%s' from the command line "+
				"arguments '%s'; expected '%s'", filename, line, expectedFileName)
		}
	}

}
