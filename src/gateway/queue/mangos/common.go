package mangos

import (
	"fmt"
	"time"
)

const (
	// How big to make the gateway/queue/mangos buffers
	channelSize = 64
	Timeout     = time.Duration(1 * time.Second)

	TCP Transport = iota
	IPC
)

// Transport represents supported Mangos transports.
type Transport int

// signal is a control signal for a close channel.
type signal struct{}

func wrapError(msg string, errOld, errNew error) error {
	// TODO: use a typed error stack implementation such as juju/errors.
	return fmt.Errorf("%s: %s: %s", msg, errNew.Error(), errOld.Error())
}
