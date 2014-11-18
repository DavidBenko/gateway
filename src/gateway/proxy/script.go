package proxy

import (
	"fmt"
	"gateway/config"
	"gateway/proxy/keys"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/robertkrimen/otto"
)

func scriptsFromFilesystem(conf config.ProxyServer) (map[string]*otto.Script, error) {
	fileInfo, err := os.Stat(conf.CodePath)

	if err != nil || (fileInfo.Mode()&os.ModeDir != os.ModeDir) {
		return nil, fmt.Errorf("Specified path to proxy code does not exist: %s\n",
			conf.CodePath)
	}

	scripts, err := buildScripts(conf.CodePath)
	if err != nil {
		return nil, fmt.Errorf("Could not build scripts: %v", err)
	}

	return scripts, nil
}

func buildScripts(codePath string) (map[string]*otto.Script, error) {
	scripts := make(map[string]*otto.Script)
	vm := otto.New()

	codePath, err := filepath.Abs(codePath)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(codePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Could not walk path: %s (%v)",
				path, err)
		}

		if isScriptFile(path) {
			key, err := keys.ScriptKeyForPaths(codePath, path)
			if err != nil {
				return fmt.Errorf("Could not create code key from path: %s (%v)",
					path, err)
			}

			scriptContents, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("Could not read script at path: %s (%v)",
					path, err)
			}

			ottoScript, err := vm.Compile(path, scriptContents)
			if err != nil {
				return fmt.Errorf("Could compile script at path: %s (%v)",
					path, err)
			}

			scripts[key] = ottoScript
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return scripts, nil
}

func isScriptFile(path string) bool {
	return filepath.Ext(path) == ".js"
}
