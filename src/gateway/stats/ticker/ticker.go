package ticker

import (
	"errors"
	"gateway/stats"
	"sort"
	"time"
)

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

	buffCh chan []stats.Point
}

// Make returns a new *Ticker and a channel to receive any errors on.  Using a
// nil backend will cause panics.
func Make(backend stats.Logger) *Ticker {
	tkr := &Ticker{
		backend: backend,
		buffCh:  make(chan []stats.Point, 1024),
	}

	return tkr
}

// Start starts the Ticker with the given control channels.
func (t *Ticker) Start(
	die <-chan struct{},
	tick <-chan time.Time,
	errCh chan<- error,
) {
	go consume(t.backend.Log, t.buffCh, tick, die, errCh)
}

// Log implements stats.Logger.Log on Ticker.  Make sure to call Start before
// calling Log, since Log will error if the Ticker is not running.
func (t *Ticker) Log(ps ...stats.Point) error {
	select {
	case t.buffCh <- ps:
		return nil
	default:
		return errors.New("tried to log to full stats buffer")
	}
}

func consume(
	log func(...stats.Point) error,
	buffCh <-chan []stats.Point,
	tick <-chan time.Time,
	die <-chan struct{},
	errCh chan<- error,
) {
	var (
		ps  []stats.Point
		buf []stats.Point

		// Note 'buf' is shadowed.
		writeBuffer = func(buf []stats.Point) {
			// Order may not be perfect, but should be good enough
			// to help an ORDER BY in the Sampler.
			sort.Sort(byDate(buf))
			if err := log(buf...); err != nil {
				errCh <- err
			}
		}
	)

	for {
		select {
		case <-die:
			close(errCh)
			return
		case <-tick:
			if buf != nil {
				go writeBuffer(buf)
			}

			buf = nil
		case ps = <-buffCh:
			buf = append(buf, ps...)
		}
	}
}
