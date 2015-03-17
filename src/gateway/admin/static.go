package admin

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"text/template"
	"time"

	"gateway/config"
	"gateway/version"

	"github.com/gorilla/mux"
)

var pathRegex = regexp.MustCompile(`API_BASE_PATH_PLACEHOLDER`)
var slashPathRegex = regexp.MustCompile(`/API_BASE_PATH_PLACEHOLDER`)

func adminStaticFileHandler(conf config.ProxyAdmin) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := mux.Vars(r)["path"]

		if path == "" || path == "index.html" {
			serveIndex(w, r, conf)
			return
		}

		serveFile(w, r, path)
	}
}

func serveFile(w http.ResponseWriter, r *http.Request, path string) {
	data, err := Asset(path)
	if err != nil || len(data) == 0 {
		http.NotFound(w, r)
		return
	}

	content := bytes.NewReader(data)
	gzipped, err := regexp.MatchString(`(js|css|svg|png|eot|ttf|woff)\z`, path)
	if gzipped {
		w.Header().Set("Content-Encoding", "gzip")
	}
	http.ServeContent(w, r, path, time.Time{}, content)
}

func serveIndex(w http.ResponseWriter, r *http.Request, conf config.ProxyAdmin) {
	data, err := Asset("index.html.template")
	if err != nil {
		http.Error(w, "Could not find index template.", http.StatusInternalServerError)
		return
	}

	funcs := template.FuncMap{
		"replacePath": func(input string) string {
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
			return pathRegex.ReplaceAllStringFunc(input, pathReplacer(clean))
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
