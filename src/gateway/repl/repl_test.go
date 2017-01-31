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
	done := make(chan error, 1)
	input := make(chan []byte)
	// All tests reuse the same REPL so set variables persist between executions
	repl, err := repl.NewRepl(otto.New(), input)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(repl, gc.NotNil)

	for i, t := range []struct {
		should       string
		givenInput   string
		expectOutput string
	}{{
		should:       "execute simple JS",
		givenInput:   "var foo = 'something';",
		expectOutput: "undefined",
	}, {
		should:       "return a value from executed JS",
		givenInput:   "'foo'",
		expectOutput: "foo",
	}, {
		should:       "return a value from a function",
		givenInput:   "function test() { return 'test'; };\ntest()",
		expectOutput: "test",
	}, {
		should:       "return values created in previous executions",
		givenInput:   "test()",
		expectOutput: "test",
	}, {
		should:       "return an error",
		givenInput:   "wrongFunc();",
		expectOutput: "error: ReferenceError: 'wrongFunc' is not defined",
	}} {
		c.Logf("Test %d: should %s", i, t.should)

		go func() {
			defer repl.Stop()
			msg := <-repl.Output
			if t.expectOutput != "" {
				c.Assert(string(msg), gc.Equals, t.expectOutput)
			}
		}()

		go func() {
			repl.Run()
			done <- nil
		}()

		if t.givenInput != "" {
			input <- []byte(t.givenInput)
		}

		<-done
		c.Assert(err, jc.ErrorIsNil)
	}
}
