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
		givenBackend     *statst.Logger
		points           []stats.Point
		expectBackendErr string
	}{{
		should:           "send any backend error back",
		givenBackend:     &statst.Logger{Error: errors.New("oops")},
		expectBackendErr: "oops",
	}, {
		should:       "log a single point",
		givenBackend: &statst.Logger{},
		points: []stats.Point{
			{Timestamp: tNow},
			{Timestamp: tNow.Add(1 * time.Second)},
		},
	}, {
		should:       "log a few points and order them",
		givenBackend: &statst.Logger{},
	}, {
		should:       "log another few",
		givenBackend: &statst.Logger{},
		points: []stats.Point{
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
		result := tkr.Log(t.points...)
		if t.expectBackendErr == "" {
			c.Assert(result, gc.IsNil)
		} else {
			c.Check(result, gc.ErrorMatches, t.expectBackendErr)
		}

	}
}
