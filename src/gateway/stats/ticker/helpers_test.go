package ticker_test

import (
	"gateway/stats"
	statst "gateway/stats/testing"
	"gateway/stats/ticker"

	gc "gopkg.in/check.v1"
)

func testLog(
	c *gc.C,
	tkr *ticker.Ticker,
	backend *statst.Logger,
	given []stats.Point,
	expectBackendErr string,
	expectLogged []stats.Point,
) {

	c.Check(backend.Buffer, gc.DeepEquals, expectLogged)

}
