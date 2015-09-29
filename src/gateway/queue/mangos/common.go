package mangos

const (
	// How big to make the gateway/queue/mangos buffers
	numChannels = 64
)

// signal is a control signal for a close channel.
type signal struct{}
