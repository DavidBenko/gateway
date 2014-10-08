package admin

import (
	"fmt"
	"net/http"
)

func adminHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "This is an admin page.\n")
}
