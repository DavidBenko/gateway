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

	"github.com/gorilla/mux"
)

var pathRegex = regexp.MustCompile(`/API_BASE_PATH_PLACEHOLDER`)

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
	http.ServeContent(w, r, path, time.Time{}, content)
}

func serveIndex(w http.ResponseWriter, r *http.Request, conf config.ProxyAdmin) {
	data, err := Asset("index.html.template")
	if err != nil {
		http.Error(w, "Could not find index template.", http.StatusInternalServerError)
		return
	}

	funcs := template.FuncMap{
		"baseHref": func() string {
			if conf.PathPrefix == "" {
				return "/"
			}
			return conf.PathPrefix
		},
		"replacePath": func(input string) string {
			return pathRegex.ReplaceAllStringFunc(input, func(string) string {
				if conf.PathPrefix == "" {
					return ""
				}

				return strings.TrimRight(conf.PathPrefix, "/")
			})
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
