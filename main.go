package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"os"

	"github.com/robertkrimen/otto"

	"github.com/AnyPresence/gateway/config"
)

func main() {

	conf, err := config.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(fmt.Sprintf("Error parsing config file: %v", err))
	}

	http.HandleFunc("/", proxyHandlerFunc)
	listen := fmt.Sprintf(":%d", conf.Proxy.Port)
	log.Fatal(http.ListenAndServe(listen, nil))
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
