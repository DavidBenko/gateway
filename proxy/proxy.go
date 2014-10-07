package proxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/AnyPresence/gateway/config"
	"github.com/AnyPresence/gateway/proxy/admin"
	"github.com/gorilla/mux"
	"github.com/robertkrimen/otto"
)

// Run the proxy server
func Run(config config.ProxyServer) {
	router := mux.NewRouter()

	// Set up admin
	admin.AddRoutes(router, config.Admin)

	// Set up proxy
	router.HandleFunc("/{path:.*}", proxyHandlerFunc).MatcherFunc(
		func(r *http.Request, rm *mux.RouteMatch) bool {
			return true
		})

	// Run server
	listen := fmt.Sprintf(":%d", config.Port)
	log.Fatal(http.ListenAndServe(listen, router))
}

func proxyHandlerFunc(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}

	req, err := http.NewRequest(r.Method, "http://localhost:4567", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Body = ioutil.NopCloser(massageBody(bytes.NewBuffer(body), `
    body = body.replace("Stan", "Kyle");
  `))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	newBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	newRespBody := massageBody(bytes.NewBuffer(newBody), `
    body = body.replace("Hello", "Bye bye");
  `)

	fmt.Fprint(w, newRespBody.String())
}

func massageBody(body *bytes.Buffer, src interface{}) *bytes.Buffer {
	vm := otto.New()
	vm.Set("body", body.String())
	vm.Run(src)
	newBodyRaw, err := vm.Get("body")
	if err != nil {
		log.Fatal(err)
	}

	newBody, err := newBodyRaw.ToString()
	if err != nil {
		log.Fatal(err)
	}

	return bytes.NewBufferString(newBody)
}
