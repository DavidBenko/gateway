package admin

import (
	"bytes"
	"fmt"
	"mime"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"text/template"
	"time"

	"gateway/config"
	"gateway/logreport"
	"gateway/model"
	"gateway/proxy/vm"
	"gateway/version"

	"github.com/gorilla/mux"
)

var pathRegex = regexp.MustCompile(`API_BASE_PATH_PLACEHOLDER`)
var slashPathRegex = regexp.MustCompile(`/API_BASE_PATH_PLACEHOLDER`)
var brokerHostRegex = regexp.MustCompile(`BROKER_PLACEHOLDER`)
var versionRegex = regexp.MustCompile(`VERSION`)
var commitRegex = regexp.MustCompile(`COMMIT`)
var devModeRegex = regexp.MustCompile(`DEV_MODE`)
var goosRegex = regexp.MustCompile(`GOOS`)
var remoteEndpointTypesEnabledRegex = regexp.MustCompile(`REMOTE_ENDPOINT_TYPES_ENABLED`)
var registrationEnabledRegex = regexp.MustCompile(`REGISTRATION_ENABLED`)

// Normalize some mime types across OSes
var additionalMimeTypes = map[string]string{
	".svg": "image/svg+xml",
}

type assetResolver func(path string) ([]byte, error)

func init() {
	for k, v := range additionalMimeTypes {
		if err := mime.AddExtensionType(k, v); err != nil {
			logreport.Fatalf("Could not set mime type for %s", k)
		}
	}
}

func adminStaticFileHandler(conf config.ProxyAdmin) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := mux.Vars(r)["path"]

		if path == "" || path == "index.html" {
			serveIndex(w, r, conf)
			return
		}

		// Make JS request objects & functions available to the front-end so that
		// the front-end can introspect on those functions for autocomplete purposes
		if path == "ap-request.js" {
			serveAsset(w, r, "http/request.js", vm.Asset)
			return
		}

		serveFile(w, r, path)
	}
}

func serveFile(w http.ResponseWriter, r *http.Request, path string) {
	serveAsset(w, r, path, Asset)
}

func serveAsset(w http.ResponseWriter, r *http.Request, path string, assetResolverFunc assetResolver) {
	data, err := assetResolverFunc(path)
	if err != nil || len(data) == 0 {
		http.NotFound(w, r)
		return
	}
	content := bytes.NewReader(data)
	http.ServeContent(w, r, path, time.Time{}, content)
}

func serveIndex(w http.ResponseWriter, r *http.Request, conf config.ProxyAdmin) {
	data, err := Asset("index.html.template")
	if err != nil {
		http.Error(w, "Could not find index template.", http.StatusInternalServerError)
		return
	}

	funcs := template.FuncMap{
		"interpolate": func(input string) string {
			pathReplacer := func(path string) func(string) string {
				return func(string) string {
					if conf.PathPrefix == "" {
						return ""
					}
					return path
				}
			}
			rightless := strings.TrimRight(conf.PathPrefix, "/")
			clean := strings.TrimLeft(rightless, "/")
			input = slashPathRegex.ReplaceAllStringFunc(input, pathReplacer(rightless))
			input = pathRegex.ReplaceAllStringFunc(input, pathReplacer(clean))

			interpolatedValues := map[*regexp.Regexp]string{}

			if conf.ShowVersion {
				interpolatedValues[versionRegex] = version.Name()
				interpolatedValues[commitRegex] = version.Commit()
			}
			interpolatedValues[devModeRegex] = fmt.Sprintf("%t", conf.DevMode)
			interpolatedValues[goosRegex] = runtime.GOOS
			interpolatedValues[remoteEndpointTypesEnabledRegex] = remoteEndpointTypes()
			interpolatedValues[registrationEnabledRegex] = fmt.Sprintf("%t", conf.EnableRegistration)
			interpolatedValues[brokerHostRegex] = conf.BrokerWs

			for k, v := range interpolatedValues {
				input = k.ReplaceAllLiteralString(input, v)
			}

			return input
		},
	}

	tmpl := template.New("index")
	if _, err = tmpl.Funcs(funcs).Parse(string(data)); err != nil {
		http.Error(w, "Could not parse index template.", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, nil)
	if err != nil {
		fmt.Fprintf(w, "\n\nError executing template: %v", err)
	}
}

func remoteEndpointTypes() string {
	tags := []string{}
	remoteEndpointTypes, _ := model.AllRemoteEndpointTypes(nil)
	for _, re := range remoteEndpointTypes {
		if re.Enabled {
			tags = append(tags, re.Value)
		}
	}
	return strings.Join(tags, ",")
}
