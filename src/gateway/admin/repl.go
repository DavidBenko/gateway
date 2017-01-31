package admin

import (
	"gateway/core"

	"golang.org/x/net/websocket"
)

type ReplController struct {
	BaseController
}

func (c *ReplController) replHandler(ws *websocket.Conn) {
	vm := core.VMCopy(c.accountID, nil)
}
