package mangos

import "time"

const (
	// How big to make the gateway/queue/mangos buffers
	channelSize = 64
	Timeout     = time.Duration(1 * time.Second)
)

// signal is a control signal for a close channel.
type signal struct{}
