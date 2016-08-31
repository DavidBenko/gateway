package ticker

import "gateway/stats"

type byDate []stats.Point

// Less implements sort.Interface on []stats.Point.
func (b byDate) Less(i, j int) bool {
	iT, jT := b[i].Timestamp, b[j].Timestamp
	return iT.Before(jT)
}

// Swap implements sort.Interface on []stats.Point.
func (b byDate) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

// Len implements sort.Interface on []stats.Point.
func (b byDate) Len() int { return len(b) }

// Ticker implements stats.Logger by collecting and dumping points to a
// stats.Logger backend.  Before dumping, it will sort the points it has
// collected by date to ease ORDER BY load on the Sampler backend.
type Ticker struct {
	backend stats.Logger

	buffer []stats.Point
}

// Make returns a new *Ticker
func Make(backend stats.Logger) *Ticker {
	tkr := &Ticker{
		backend: backend,
		buffer:  make([]stats.Point, 1024),
	}

	return tkr
}

// Log implements stats.Logger.Log on Ticker.  Make sure to call Start before
// calling Log, since Log will error if the Ticker is not running.
func (t *Ticker) Log(ps ...stats.Point) error {
	return t.backend.Log(ps...)
}
