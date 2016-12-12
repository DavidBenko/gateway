package admin

import (
	"net"
	"net/http"
	"net/url"

	"gateway/config"
	aphttp "gateway/http"
	"gateway/logreport"

	"github.com/gorilla/handlers"
	"golang.org/x/net/websocket"
)

type MQTTProxyController struct {
	BaseController
	config.Push
}

func RouteMQTTProxy(c *MQTTProxyController, path string, router aphttp.Router, conf config.ProxyAdmin) {
	server := websocket.Server{
		Config: websocket.Config{
			Version: websocket.ProtocolVersionHybi13,
		},
		Handler: websocket.Handler(c.handler),
	}
	routes := map[string]http.Handler{
		"GET": http.HandlerFunc(server.ServeHTTP),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "OPTIONS"})
	}
	router.Handle(path, handlers.MethodHandler(routes))
}

func (c *MQTTProxyController) handler(ws *websocket.Conn) {
	logError := func(err error) {
		logreport.Printf("[mqtt_proxy] %v", err)
	}

	defer ws.Close()

	uri, err := url.Parse(c.MQTTURI)
	if err != nil {
		logError(err)
		return
	}
	client, err := net.Dial(uri.Scheme, uri.Host)
	if err != nil {
		logError(err)
		return
	}
	defer client.Close()

	done := make(chan bool)

	// send
	go func() {
		defer func() {
			done <- true
		}()

		buffer := make([]byte, 2048)
		for {
			n, err := client.Read(buffer)
			if err != nil || n < 1 {
				logError(err)
				return
			}
			err = websocket.Message.Send(ws, buffer[:n])
			if err != nil {
				logError(err)
				return
			}
		}
	}()

	// receive
	go func() {
		defer func() {
			done <- true
		}()

		var buffer []byte
		for {
			err := websocket.Message.Receive(ws, &buffer)
			if err != nil {
				logError(err)
				return
			}
			n, err := client.Write(buffer)
			if err != nil || n < 1 {
				logError(err)
				return
			}
		}
	}()

	<-done
}
