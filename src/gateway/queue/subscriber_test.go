package queue_test

import (
	"gateway/queue"
	qt "gateway/queue/testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func (s *QueueSuite) TestSubscribe(c *gc.C) {
	for i, t := range []struct {
		should    string
		path      string
		bindings  []queue.SubBinding
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
		bindings: []queue.SubBinding{
			qt.SubBindingOk,
			qt.SubBindingErr("an error"),
		},
		expectErr: "bad subscriber binding: an error",
	}, {
		should: "error if Connect() errors",
		path:   "foo",
		bindings: []queue.SubBinding{
			qt.SubBindingOk,
			qt.SubBindingConnectErr("connect error"),
		},
		expectErr: "subscriber failed to connect: connect error",
	}, {
		should: "work if there are no errors",
		path:   "foo",
		bindings: []queue.SubBinding{
			qt.SubBindingOk,
		},
	}} {
		c.Logf("test %d: should %s", i, t.should)

		s, err := queue.Subscribe(t.path, t.bindings...)
		if t.expectErr != "" {
			c.Check(err, gc.ErrorMatches, t.expectErr)
			continue
		}

		c.Check(err, jc.ErrorIsNil)
		c.Check(s, gc.NotNil)
	}
}
