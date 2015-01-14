package admin

import (
	"bytes"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func adminStaticFileHandler(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	if path == "" {
		path = "index.html"
	}
	serveFile(w, r, path)
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
