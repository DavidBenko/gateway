package admin

import (
	"fmt"
	"net"
	"net/url"

	"gateway/config"
	aphttp "gateway/http"
	"gateway/logreport"

	"golang.org/x/net/websocket"
)

type MQTTProxyController struct {
	BaseController
	config.Push
}

func RouteMQTTProxy(c *MQTTProxyController, path string, router aphttp.Router) {
	router.Handle(path, websocket.Handler(c.handler))
}

func (c *MQTTProxyController) handler(ws *websocket.Conn) {
	fmt.Println("handler")
	logError := func(err error) {
		logreport.Printf("[mqtt_proxy] %v", err)
	}

	defer ws.Close()
	fmt.Println("start")
	uri, err := url.Parse(c.MQTTURI)
	if err != nil {
		logError(err)
		return
	}
	fmt.Println("here")
	client, err := net.Dial(uri.Scheme, uri.Host)
	if err != nil {
		logError(err)
		return
	}
	defer client.Close()
	fmt.Println("there")
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
	fmt.Println("done")
	<-done
}
