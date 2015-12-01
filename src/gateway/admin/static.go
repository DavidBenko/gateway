package admin

import (
	"bytes"
	"fmt"
	"log"
	"mime"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"text/template"
	"time"

	"gateway/config"
	"gateway/model"
	"gateway/proxy/vm"
	"gateway/version"

	"github.com/gorilla/mux"
)

var pathRegex = regexp.MustCompile(`API_BASE_PATH_PLACEHOLDER`)
var slashPathRegex = regexp.MustCompile(`/API_BASE_PATH_PLACEHOLDER`)
var brokerHostRegex = regexp.MustCompile(`BROKER_PLACEHOLDER`)

// Normalize some mime types across OSes
var additionalMimeTypes = map[string]string{
	".svg": "image/svg+xml",
}

type assetResolver func(path string) ([]byte, error)

func init() {
	for k, v := range additionalMimeTypes {
		if err := mime.AddExtensionType(k, v); err != nil {
			log.Fatalf("Could not set mime type for %s", k)
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

			input = brokerHostRegex.ReplaceAllLiteralString(input, conf.Broker)

			return input
		},
		"version": func() string {
			if !conf.ShowVersion {
				return ""
			}

			return fmt.Sprintf("<meta name=\"version\" content=\"%s\">\n<meta name=\"commit\" content=\"%s\">",
				version.Name(), version.Commit())
		},
		"devMode": func() string {
			if !conf.DevMode {
				return ""
			}

			return "<meta name=\"dev-mode\" content=\"true\">"
		},
		"goos": func() string {
			return fmt.Sprintf("<meta name=\"goos\" content=\"%s\">", runtime.GOOS)
		},
		"remoteEndpointTypes": func() string {
			tags := []string{}
			remoteEndpointTypes, _ := model.AllRemoteEndpointTypes(nil)
			for _, re := range remoteEndpointTypes {
				if re.Enabled {
					tags = append(tags, re.Value)
				}
			}
			return fmt.Sprintf("<meta name=\"remote-endoint-types-enabled\" content=\"%s\">", strings.Join(tags, ","))
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
