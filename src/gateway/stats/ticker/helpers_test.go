package ticker_test

import (
	"math/rand"
	"runtime"
	"time"

	"gateway/stats"
	statst "gateway/stats/testing"
	"gateway/stats/ticker"

	gc "gopkg.in/check.v1"
)

func testLog(
	c *gc.C,
	tkr *ticker.Ticker,
	die chan<- struct{},
	ctrl chan time.Time,
	BEErr chan error,
	backend *statst.Logger,
	givenFirst, givenNext [][]stats.Point,
	expectBackendErr string,
	expectLoggedFirst, expectLoggedNext []stats.Point,
) {
	defer func() {
		// Don't forget to clean up!
		close(die)
	}()

	if !checkLogAll(c, tkr, givenFirst, "") {
		return
	}

	if !checkNotYetTriggered(c, backend.Buffer, BEErr) {
		return
	}

	// Tick once.
	ctrl <- time.Now()

	// The first set of points should now have been logged, unless there was
	// a backend error.
	checkErrCh(c, BEErr, expectBackendErr)
	if expectBackendErr != "" {
		return
	}

	c.Check(backend.Buffer, gc.DeepEquals, expectLoggedFirst)

	if givenNext == nil {
		// If we were only testing one set of points, we're done.
		return
	}

	// If we haven't returned yet, we're testing for more good behavior. Log
	// the next set of points.
	if !checkLogAll(c, tkr, givenNext, "") {
		return
	}

	// Make sure we have still only seen the first set of points.
	c.Check(backend.Buffer, gc.DeepEquals, expectLoggedFirst)

	// Tick once.
	ctrl <- time.Now()

	// Now all points should have been logged, with no error.
	checkErrCh(c, BEErr, "")
	c.Check(backend.Buffer, gc.DeepEquals, expectLoggedNext)
}

func checkLogAll(
	c *gc.C,
	tkr *ticker.Ticker,
	given [][]stats.Point,
	err string,
) bool {
	// Randomize order to simulate concurrent access.
	for _, i := range rand.Perm(len(given)) {
		if !checkLog(c, tkr, given[i], err) {
			return false
		}
	}

	return true
}

// false: don't continue this test
func checkLog(c *gc.C, tkr *ticker.Ticker, ps []stats.Point, expect string) bool {
	err := tkr.Log(ps...)

	if expect != "" {
		c.Check(err, gc.ErrorMatches, expect)
		return false
	} else {
		c.Assert(err, gc.IsNil)
		return true
	}
}

// false: don't continue this test
func checkNotYetTriggered(c *gc.C, buf []stats.Point, err chan error) bool {
	c.Assert(buf, gc.IsNil)
	checkErrCh(c, err, "")

	return !c.Failed()
}

func checkErrCh(c *gc.C, err chan error, expect string) {
	runtime.Gosched()

	var e error

	select {
	case e = <-err:
	default:
	}

	if expect != "" {
		c.Check(e, gc.ErrorMatches, expect)
		return
	}

	c.Assert(e, gc.IsNil)
}
