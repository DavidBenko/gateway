package mangos

import (
	"fmt"
	"time"
)

const (
	// How big to make the gateway/queue/mangos buffers
	channelSize = 64
	// Timeout is how long to wait to clean up PubSockets and SubSockets.
	Timeout = time.Duration(1 * time.Second)

	// TCP is a transport over IP.
	TCP Transport = iota
	// IPC is a machine-local transport not supported on Windows.
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
