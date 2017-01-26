package repl_test

import (
	"gateway/repl"
	"io"
	"testing"

	"golang.org/x/net/websocket"

	"github.com/robertkrimen/otto"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) { gc.TestingT(t) }

type ReplSuite struct{}

var _ = gc.Suite(&ReplSuite{})

func (s *ReplSuite) TestNewRepl(c *gc.C) {
	for i, t := range []struct {
		should         string
		givenVM        *otto.Otto
		givenRWC       io.ReadWriteCloser
		givenAccountID int64
		expectError    string
	}{{
		should:         "accept a websocket to satisify the RWC interface",
		givenVM:        nil,
		givenRWC:       &websocket.Conn{},
		givenAccountID: 1,
	}, {
		should:         "return an error for a 0 accountID",
		givenVM:        nil,
		givenRWC:       &websocket.Conn{},
		givenAccountID: 0,
		expectError:    "invalid accountID 0",
	}} {

		c.Logf("Test %d: should %s", i, t.should)

		given, err := repl.NewRepl(t.givenVM, t.givenRWC, t.givenAccountID)
		if t.expectError != "" {
			c.Assert(err.Error(), gc.Equals, t.expectError)
			break
		}
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(given, gc.NotNil)
	}
}
