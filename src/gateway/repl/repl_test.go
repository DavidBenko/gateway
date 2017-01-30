package repl_test

import (
	"gateway/repl"
	"testing"

	"github.com/robertkrimen/otto"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) { gc.TestingT(t) }

type ReplSuite struct{}

var _ = gc.Suite(&ReplSuite{})

func (s *ReplSuite) TestRepl(c *gc.C) {
	for i, t := range []struct {
		should       string
		givenInput   string
		expectOutput string
		expectError  string
	}{{
		should:       "get welcome message",
		expectOutput: "Hello",
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		input := make(chan []byte)
		repl, err := repl.NewRepl(otto.New(), input)
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(repl, gc.NotNil)
		c.Assert(repl.Output, gc.NotNil)

		go func() {
			select {
			case msg := <-repl.Output:

			}
		}()

		repl.Start()
	}
}
