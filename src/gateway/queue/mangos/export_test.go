package mangos

import (
	"gateway/queue"

	"github.com/gdamore/mangos"
)

func GetPubSocket(p queue.Publisher) (mangos.Socket, error) {
	return getPubSocket(p)
}

func GetSubSocket(p queue.Subscriber) (mangos.Socket, error) {
	return getSubSocket(p)
}

func GetPubBufferSize(p queue.Publisher) int {
	tP := p.(*PubSocket)
	return tP.buffSize
}

func GetSubBufferSize(s queue.Subscriber) int {
	tS := s.(*SubSocket)
	return tS.buffSize
}
