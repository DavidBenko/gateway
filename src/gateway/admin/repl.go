package admin

import "golang.org/x/net/websocket"

type ReplController struct {
	BaseController
}

func (c *ReplController) replHandler(ws *websocket.Conn) {
}
