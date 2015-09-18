package queue_test

import (
	"gateway/queue"
	qt "gateway/queue/testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func (s *QueueSuite) TestPublish(c *gc.C) {
	for i, t := range []struct {
		should    string
		path      string
		bindings  []queue.PubBinding
		expectErr string
	}{{
		should:    "error with no path",
		expectErr: "no path provided",
	}, {
		should:    "error with no bindings",
		path:      "foo",
		expectErr: "no bindings provided",
	}, {
		should: "error if a binding returns an error",
		path:   "foo",
		bindings: []queue.PubBinding{
			qt.PubBindingOk,
			qt.PubBindingErr("an error"),
		},
		expectErr: "bad publisher binding: an error",
	}, {
		should: "error if Bind() errors",
		path:   "foo",
		bindings: []queue.PubBinding{
			qt.PubBindingOk,
			qt.PubBindingBindErr("bind error"),
		},
		expectErr: "publisher failed to bind: bind error",
	}, {
		should: "work if there are no errors",
		path:   "foo",
		bindings: []queue.PubBinding{
			qt.PubBindingOk,
		},
	}} {
		c.Logf("test %d: should %s", i, t.should)

		p, err := queue.Publish(t.path, t.bindings...)

		if t.expectErr != "" {
			c.Check(err, gc.ErrorMatches, t.expectErr)
			continue
		}

		c.Assert(err, jc.ErrorIsNil)
		c.Assert(p, gc.NotNil)
	}
}
