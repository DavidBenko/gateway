package mangos

import (
	"gateway/queue"
	"reflect"

	"github.com/gdamore/mangos"
	gc "gopkg.in/check.v1"
)

func GetPubSocket(p queue.Publisher) (mangos.Socket, error) {
	return getPubSocket(p)
}

func GetSubSocket(p queue.Subscriber) (mangos.Socket, error) {
	return getSubSocket(p)
}

// GetPubBufferSize returns the size of the internal mangos buffer.
func GetPubBufferSize(p queue.Publisher) int {
	tP := p.(*PubSocket)
	l, err := tP.s.GetOption(mangos.OptionWriteQLen)
	if err != nil {
		panic(err)
	}

	return l.(int)
}

// GetSubBufferSize returns the size of the internal mangos buffer.
func GetSubBufferSize(s queue.Subscriber) int {
	tS := s.(*SubSocket)
	l, err := tS.s.GetOption(mangos.OptionReadQLen)
	if err != nil {
		panic(err)
	}

	return l.(int)
}

func IsBrokered(c *gc.C, p queue.Publisher) bool {
	c.Assert(reflect.TypeOf(p), gc.Equals, reflect.TypeOf(&PubSocket{}))
	ps := p.(*PubSocket)
	return ps.useBroker
}
