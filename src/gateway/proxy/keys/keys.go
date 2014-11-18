package keys

import (
	"path/filepath"
	"strings"
)

// ScriptKeyForPaths turns a path into a key to use to
// canonically identify a script
func ScriptKeyForPaths(codePath, scriptPath string) (string, error) {
	key, err := filepath.Rel(codePath, scriptPath)
	if err != nil {
		return "", err
	}
	key = strings.TrimSuffix(key, ".js")
	key = strings.Replace(key, string(filepath.Separator), ".", -1)
	key = strings.Replace(key, "-", "", -1)
	key = strings.Replace(key, "_", "", -1)
	key = strings.ToLower(key)
	return key, nil
}

// ScriptKeyForCodeString returns a canonical script key from a string
// that is equivalent to what would be used in code.
func ScriptKeyForCodeString(code string) (string, error) {
	key := strings.ToLower(code)
	key = strings.Replace(key, string(filepath.Separator), ".", -1)
	return key, nil
}
