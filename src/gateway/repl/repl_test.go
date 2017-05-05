package repl_test

import (
	"gateway/repl"
	"testing"

	"github.com/robertkrimen/otto"

	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) { gc.TestingT(t) }

type ReplSuite struct{}

var _ = gc.Suite(&ReplSuite{})

func (s *ReplSuite) TestRepl(c *gc.C) {
	done := make(chan error, 1)
	// All tests reuse the same REPL so set variables persist between executions
	r := repl.NewRepl(otto.New())
	c.Assert(r, gc.NotNil)

	for i, t := range []struct {
		should       string
		givenInput   string
		expectOutput *repl.Frame
	}{{
		should:       "execute simple JS",
		givenInput:   "var foo = 'something';",
		expectOutput: &repl.Frame{Data: "undefined", Type: "output"},
	}, {
		should:       "return a value from executed JS",
		givenInput:   "'foo'",
		expectOutput: &repl.Frame{Data: "foo", Type: "output"},
	}, {
		should:       "return a value from a function",
		givenInput:   "function test() { return 'test'; };\ntest()",
		expectOutput: &repl.Frame{Data: "test", Type: "output"},
	}, {
		should:       "return values created in previous executions",
		givenInput:   "test()",
		expectOutput: &repl.Frame{Data: "test", Type: "output"},
	}, {
		should:       "return an error",
		givenInput:   "wrongFunc();",
		expectOutput: &repl.Frame{Data: "ReferenceError: 'wrongFunc' is not defined", Type: "error"},
	}} {
		c.Logf("Test %d: should %s", i, t.should)

		go func() {
			defer r.Stop()
			msg := <-r.Output
			if t.expectOutput != nil {
				c.Assert(string(msg), gc.Equals, string(t.expectOutput.JSON()))
			}
		}()

		go func() {
			r.Run()
			done <- nil
		}()

		if t.givenInput != "" {
			r.Input <- []byte(t.givenInput)
		}

		<-done
	}
}
