package proxy

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/robertkrimen/otto"
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

func testSimpleCases(t *testing.T, router *mux.Router,
	testCases map[string]string) {
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

func TestSchemes(t *testing.T) {
	router := buildRouter(t, `
		router.get("/foo", "a").scheme("http");
		router.get("/foo", "b").scheme("https");
		router.get("/bar", "c").schemes("http", "https");
	`)
	testURLCases(t, router, map[string]string{
		"http://www.example.com/foo":  "a",
		"https://www.example.com/foo": "b",
		"http://www.example.com/bar":  "c",
		"https://www.example.com/bar": "c",
	})
}

func TestOnlySchemes(t *testing.T) {
	router := buildRouter(t, `
		router.scheme("http").name("a");
		router.scheme("https").name("b");
	`)
	testURLCases(t, router, map[string]string{
		"http://www.example.com":  "a",
		"https://www.example.com": "b",
	})
}

func TestHosts(t *testing.T) {
	router := buildRouter(t, `
		router.get("/foo", "a").host("www.example.com");
		router.get("/foo", "b").host("subdomain.example.com");
	`)
	testURLCases(t, router, map[string]string{
		"http://www.example.com/foo":        "a",
		"https://subdomain.example.com/foo": "b",
		"https://nope.example.com/bar":      "",
	})
}

func TestOnlyHosts(t *testing.T) {
	router := buildRouter(t, `
		router.host("www.example.com").name("a");
		router.host("subdomain.example.com").name("b");
	`)
	testURLCases(t, router, map[string]string{
		"http://www.example.com/foo":       "a",
		"http://subdomain.example.com/bar": "b",
		"https://nope.example.com/bar":     "",
	})
}

func TestPathPrefix(t *testing.T) {
	router := buildRouter(t, `
		router.pathPrefix("/a").name("a");
		router.pathPrefix("/b").name("b");
	`)
	testURLCases(t, router, map[string]string{
		"http://www.example.com/alpha":   "a",
		"http://www.example.com/acrobat": "a",
		"http://www.example.com/beta":    "b",
		"http://www.example.com/bbq":     "b",
		"http://www.example.com/kappa":   "",
	})
}

func testURLCases(t *testing.T, router *mux.Router,
	testCases map[string]string) {
	for url, name := range testCases {
		request, _ := http.NewRequest("GET", url, nil)
		var match mux.RouteMatch
		ok := router.Match(request, &match)

		if name == "" {
			if ok {
				t.Errorf("Expected %s not to be routed", url)
			}
			continue
		}

		if ok {
			if match.Route.GetName() != name {
				t.Errorf("Expected %s to route to %s, got %s",
					url, name, match.Route.GetName())
			}
		} else {
			t.Errorf("Expected %s to be routed", url)
		}
	}
}

func TestBasicRegexHost(t *testing.T) {
	router := buildRouter(t, `
		router.host("{subdomain}.example.com").name("a");
	`)
	url := "http://alpha.example.com/foo"
	request, _ := http.NewRequest("GET", url, nil)
	var match mux.RouteMatch
	if ok := router.Match(request, &match); ok {
		if match.Route.GetName() != "a" {
			t.Errorf("Expected %s to route to 'a', got %s",
				url, match.Route.GetName())
		}
		if match.Vars["subdomain"] != "alpha" {
			t.Errorf(
				"Expected route to parse subdomain var as 'alpha', got '%s'",
				match.Vars["subdomain"])
		}
	} else {
		t.Errorf("Expected %s to be routed", url)
	}
}

func TestSpecificRegexHost(t *testing.T) {
	router := buildRouter(t, `
		router.host("{subdomain:[0-9]+}.example.com").name("a");
	`)

	url := "http://111.example.com"
	request, _ := http.NewRequest("GET", url, nil)

	var match mux.RouteMatch
	if ok := router.Match(request, &match); ok {
		if match.Route.GetName() != "a" {
			t.Errorf("Expected %s to route to 'a', got %s",
				url, match.Route.GetName())
		}
		if match.Vars["subdomain"] != "111" {
			t.Errorf(
				"Expected route to parse subdomain var as 'alpha', got '%s'",
				match.Vars["subdomain"])
		}
	} else {
		t.Errorf("Expected %s to be routed", url)
	}

	url = "http://alpha.example.com"
	request, _ = http.NewRequest("GET", url, nil)
	if ok := router.Match(request, &match); ok {
		t.Errorf("Expected GET 'http://alpha.example.com' not to match")
	}
}

func TestHeaders(t *testing.T) {
	router := buildRouter(t, `
		router.headers({"Content-Type": "application/json"}).name("json");
		router.headers({"Content-Type": "application/xml"}).name("xml");
		router.headers({"Content-Type": "application/yml",
		                "X-Seriously": "Yes"}).name("yml");
	`)
	testCases := map[string]map[string]string{
		"json": map[string]string{"Content-Type": "application/json"},
		"xml":  map[string]string{"Content-Type": "application/xml"},
		"yml": map[string]string{"Content-Type": "application/yml",
			"X-Seriously": "Yes"},
		"": map[string]string{"Content-Type": "application/yml",
			"X-Seriously": "No"},
	}

	for name, headers := range testCases {
		request, _ := http.NewRequest("GET", "/foo", nil)
		for k, v := range headers {
			request.Header.Add(k, v)
		}

		var match mux.RouteMatch
		ok := router.Match(request, &match)

		if name == "" {
			if ok {
				t.Errorf("Expected headers %v not to be routed", headers)
			}
			continue
		}

		if ok {
			if match.Route.GetName() != name {
				t.Errorf("Expected headers %v to route to %s, got %s",
					headers, name, match.Route.GetName())
			}
		} else {
			t.Errorf("Expected headers %v to be routed", headers)
		}
	}
}

func TestRegexQueries(t *testing.T) {
	router := buildRouter(t, `
		router.queries({"foo": "{id}"}).name("a");
		router.queries({"bar": "{id:[0-9]+}"}).name("b");
	`)
	testCases := map[string][]string{
		"http://www.example.com?foo=a": []string{"a", "a"},
		"http://www.example.com?foo=b": []string{"a", "b"},
		"http://www.example.com?bar=1": []string{"b", "1"},
		"http://www.example.com?bar=2": []string{"b", "2"},
		"http://www.example.com?bar=a": []string{"", ""},
	}
	for url, expected := range testCases {
		request, _ := http.NewRequest("GET", url, nil)
		var match mux.RouteMatch
		ok := router.Match(request, &match)

		name := expected[0]
		id := expected[1]

		if name == "" {
			if ok {
				t.Errorf("Expected %s not to be routed", url)
			}
			continue
		}

		if ok {
			if match.Route.GetName() != name {
				t.Errorf("Expected %s to route to %s, got %s",
					url, name, match.Route.GetName())
			}
			if match.Vars["id"] != id {
				t.Errorf("Expected %s to parse id as %s, got %s",
					url, id, match.Vars["id"])
			}
		} else {
			t.Errorf("Expected %s to be routed", url)
		}
	}
}

func buildRouter(t *testing.T, script string) *mux.Router {
	vm := otto.New()
	ottoScript, err := vm.Compile("", script)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	router, err := ParseRoutes(ottoScript)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return router
}

func buildRequest(method string, path string) *http.Request {
	request, _ := http.NewRequest(method, path, nil)
	return request
}
