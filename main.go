package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	listen = flag.String("listen", ":8080", "address to listen on")
)

func main() {
	http.HandleFunc("/", proxyHandlerFunc)
	log.Fatal(http.ListenAndServe(*listen, nil))
}

func proxyHandlerFunc(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}

	req, err := http.NewRequest(r.Method, "http://localhost:4567", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Body = r.Body

	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(w, string(body))
}
