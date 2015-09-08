package queue_test

import (
	"gateway/queue"
	qt "gateway/queue/testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func (s *QueueSuite) TestSubscribe(c *gc.C) {
	reply := make(chan struct{})

	for i, t := range []struct {
		should      string
		path        string
		bindings    []queue.SubBinding
		expectErr   string
		expectClose bool
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
			qt.SubBindingCloseChan(reply),
		},
		expectErr:   "subscriber failed to connect: connect error",
		expectClose: true,
	}, {
		should: "work if there are no errors",
		path:   "foo",
		bindings: []queue.SubBinding{
			qt.SubBindingOk,
			qt.SubBindingCloseChan(reply),
		},
		expectClose: true,
	}} {
		msg := t.should
		if t.expectClose {
			msg += ", expect Close()"
		}

		c.Logf("test %d: should %s", i, msg)

		ch := testSubscribe(c, t.path, t.expectErr, t.bindings)
		_ = ch
	}
}

func testSubscribe(c *gc.C, path, errText string, bindings []queue.SubBinding) *queue.SubChannel {
	p, err := queue.Subscribe(path, bindings...)
	if errText != "" {
		c.Assert(err, gc.ErrorMatches, errText)
		return nil
	}
	c.Check(err, jc.ErrorIsNil)
	c.Check(p, gc.NotNil)

	return p
}
