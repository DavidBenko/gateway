package ticker_test

import (
	"errors"
	"testing"
	"time"

	"gateway/stats"
	statst "gateway/stats/testing"
	"gateway/stats/ticker"

	gc "gopkg.in/check.v1"
)

var (
	_ = stats.Logger(&ticker.Ticker{})
	_ = gc.Suite(&TickerSuite{})
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gc.TestingT(t) }

type TickerSuite struct{}

func (t *TickerSuite) TestLog(c *gc.C) {
	tNow := time.Now().UTC()

	for i, t := range []struct {
		should           string
		given            [][]stats.Point
		givenNext        [][]stats.Point
		givenFreq        time.Duration
		givenBackend     *statst.Logger
		expectLogged     []stats.Point
		expectLoggedNext []stats.Point
		expectBackendErr string
	}{{
		should:           "send any backend error down the channel",
		given:            [][]stats.Point{{{}}},
		givenBackend:     &statst.Logger{Error: errors.New("oops")},
		expectBackendErr: "oops",
	}, {
		should:       "log a single point after ticking",
		given:        [][]stats.Point{{{Timestamp: tNow}}},
		givenBackend: &statst.Logger{},
		expectLogged: []stats.Point{{Timestamp: tNow}},
	}, {
		should: "log a single point, tick, and log another",
		given:  [][]stats.Point{{{Timestamp: tNow}}},
		givenNext: [][]stats.Point{{
			{Timestamp: tNow.Add(1 * time.Second)},
		}},
		givenBackend: &statst.Logger{},
		expectLogged: []stats.Point{{Timestamp: tNow}},
		expectLoggedNext: []stats.Point{
			{Timestamp: tNow},
			{Timestamp: tNow.Add(1 * time.Second)},
		},
	}, {
		should: "log a few points and order them",
		given: [][]stats.Point{{
			{Timestamp: tNow.Add(2 * time.Second)},
			{Timestamp: tNow},
			{Timestamp: tNow.Add(1 * time.Second)},
		}},
		givenBackend: &statst.Logger{},
		expectLogged: []stats.Point{
			{Timestamp: tNow},
			{Timestamp: tNow.Add(1 * time.Second)},
			{Timestamp: tNow.Add(2 * time.Second)},
		},
	}, {
		should: "log a few points, tick, and log another few",
		given: [][]stats.Point{{
			{Timestamp: tNow.Add(2 * time.Second)},
			{Timestamp: tNow},
			{Timestamp: tNow.Add(1 * time.Second)},
		}},
		givenNext: [][]stats.Point{{
			{Timestamp: tNow.Add(4 * time.Second)},
			{Timestamp: tNow},
			{Timestamp: tNow.Add(1 * time.Second)}}},
		givenBackend: &statst.Logger{},
		expectLogged: []stats.Point{
			{Timestamp: tNow},
			{Timestamp: tNow.Add(1 * time.Second)},
			{Timestamp: tNow.Add(2 * time.Second)},
		},
		expectLoggedNext: []stats.Point{
			{Timestamp: tNow},
			{Timestamp: tNow.Add(1 * time.Second)},
			{Timestamp: tNow.Add(2 * time.Second)},
			{Timestamp: tNow},
			{Timestamp: tNow.Add(1 * time.Second)},
			{Timestamp: tNow.Add(4 * time.Second)},
		},
	}} {
		c.Logf("test %d: should %s", i, t.should)

		tkr := ticker.Make(t.givenBackend)
		ctrl, BEErr := make(chan time.Time), make(chan error)
		die := make(chan struct{})
		tkr.Start(die, ctrl, BEErr)

		testLog(
			c,
			tkr,
			die,
			ctrl,
			BEErr,
			t.givenBackend,
			t.given, t.givenNext,
			t.expectBackendErr,
			t.expectLogged, t.expectLoggedNext,
		)
	}
}
