package model

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
)

func TestBasicMethods(t *testing.T) {
	router := buildRouter(t, `
		router.path("/foo").method("GET").name("a");
		router.path("/foo").method("POST").name("b");
		router.path("/foo").method("PUT").name("c");
		router.path("/foo").method("PATCH").name("d");
		router.path("/foo").method("DELETE").name("e");
	`)
	testSimpleCases(t, router, map[string]string{
		"GET":    "a",
		"POST":   "b",
		"PUT":    "c",
		"PATCH":  "d",
		"DELETE": "e",
	})
}

func TestShortcuts(t *testing.T) {
	router := buildRouter(t, `
		router.get("/foo", "a");
		router.post("/foo", "b");
		router.put("/foo", "c");
		router.patch("/foo", "d");
		router.delete("/foo", "e");
  `)
	testSimpleCases(t, router, map[string]string{
		"GET":    "a",
		"POST":   "b",
		"PUT":    "c",
		"PATCH":  "d",
		"DELETE": "e",
	})
}

func TestMultipleMethods(t *testing.T) {
	router := buildRouter(t, `
		router.path("/foo").methods("GET", "POST").name("a");
	`)
	testSimpleCases(t, router, map[string]string{
		"GET":  "a",
		"POST": "a",
	})
}

func testSimpleCases(t *testing.T, router *mux.Router, testCases map[string]string) {
	for method, name := range testCases {
		request := buildRequest(method, "/foo")
		var match mux.RouteMatch
		ok := router.Match(request, &match)
		if ok {
			if match.Route.GetName() != name {
				t.Errorf("Expected matched route to be '%s'", name)
			}
		} else {
			t.Errorf("Expected %s '/foo' to match", method)
		}

	}
}

func TestBasicRegexPath(t *testing.T) {
	router := buildRouter(t, `
		router.get("/foos/{id}", "a");
	`)
	request := buildRequest("GET", "/foos/1")

	var match mux.RouteMatch
	ok := router.Match(request, &match)
	if ok {
		if match.Vars["id"] != "1" {
			t.Errorf("Expected route to parse id var as 1, got '%s'",
				match.Vars["id"])
		}
	} else {
		t.Errorf("Expected GET '/foos/1' to match")
	}
}

func TestSpecificRegexPath(t *testing.T) {
	router := buildRouter(t, `
		router.get("/foos/{id:[0-9]+}", "a");
	`)
	request := buildRequest("GET", "/foos/17")
	var match mux.RouteMatch
	ok := router.Match(request, &match)
	if ok {
		if match.Vars["id"] != "17" {
			t.Errorf("Expected route to parse id var as 17, got '%s'",
				match.Vars["id"])
		}
	} else {
		t.Errorf("Expected GET '/foos/17' to match")
	}

	request = buildRequest("GET", "/foos/alpha")
	if ok = router.Match(request, &match); ok {
		t.Errorf("Expected GET '/foos/alpha' not to match")
	}
}

func buildRouter(t *testing.T, script string) *mux.Router {
	r := Router{Script: script}
	if err := r.ParseRoutes(); err != nil {
		t.Error(err)
		t.FailNow()
	}
	return r.MUXRouter
}

func buildRequest(method string, path string) *http.Request {
	request, _ := http.NewRequest(method, path, nil)
	return request
}
