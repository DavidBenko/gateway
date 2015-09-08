package queue_test

import (
	"gateway/queue"
	qt "gateway/queue/testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func (s *QueueSuite) TestPublisher(c *gc.C) {
	var pubChan chan<- []byte
	reply := new(chan struct{})
	messageQueue := new([][]byte)

	for i, t := range []struct {
		should      string
		path        string
		bindings    []queue.PubBinding
		messages    [][]byte
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
			qt.PubBindingCloseChan(reply),
		},
		expectErr:   "publisher failed to bind: bind error",
		expectClose: true,
	}, {
		should: "work if there are no errors",
		path:   "foo",
		bindings: []queue.PubBinding{
			qt.PubBindingOk,
			qt.PubBindingBindMessages(messageQueue),
			qt.PubBindingCloseChan(reply),
		},
		messages:    qt.MakeMessages("A", "B", "C"),
		expectClose: true,
	}} {
		msg := t.should
		if t.expectClose {
			msg += ", expect Close()"
		}

		c.Logf("test %d: should %s", i, msg)

		// test manually closing
		func() {
			pubChan = nil
			*messageQueue = make([][]byte, 0)
			*reply = make(chan struct{})

			pc := testPublish(c, t.path, t.bindings, t.messages, t.expectErr, messageQueue, reply)

			// We won't test manually closing for errored calls.
			// They can get cleaned up on their own.
			if t.expectClose && t.expectErr == "" {
				c.Assert(pc, gc.NotNil)

				pubChan = pc.C

				// Shouldn't be closed yet
				c.Check(qt.IsReplyClosed(*reply), gc.Equals, false)
				c.Check(qt.IsPubChanClosed(pubChan), gc.Equals, false)

				pc.Close()
				c.Check(qt.IsPubChanClosed(pubChan), gc.Equals, true)
			}
		}()

		if !t.expectClose {
			c.Check(qt.IsReplyClosed(*reply), gc.Equals, false)
			continue
		}

		// Close signal should have arrived either way now.
		c.Assert(qt.IsReplyClosed(*reply), gc.Equals, true)

		// close over to cause pc to go out of scope.
		func() {
			qt.SyncGC()
			*messageQueue = make([][]byte, 0)
			*reply = make(chan struct{})

			pc := testPublish(c, t.path, t.bindings, t.messages, t.expectErr, messageQueue, reply)

			if t.expectErr == "" {
				c.Assert(pc, gc.NotNil)
				pubChan = pc.C
				c.Assert(qt.IsReplyClosed(*reply), gc.Equals, false)
				c.Assert(qt.IsPubChanClosed(pubChan), gc.Equals, false)
				c.Check(pc, gc.NotNil)
			} else {
				pubChan = nil
				// in case of error, should already be cleaned up
				c.Assert(qt.IsReplyClosed(*reply), gc.Equals, true)
			}
		}()

		// If there was an error, it already got cleaned up and checked
		if t.expectErr == "" {
			c.Check(qt.IsPubChanClosed(pubChan), gc.Equals, true)
			c.Check(qt.IsReplyClosed(*reply), gc.Equals, true)
		}
	}
}

func testPublish(
	c *gc.C,
	path string,
	bindings []queue.PubBinding,
	messages [][]byte,
	expectErr string,
	messageQueue *[][]byte,
	reply *chan struct{},
) *queue.PubChannel {
	p, err := queue.Publish(path, bindings...)

	if expectErr != "" {
		c.Assert(p, gc.IsNil)
		c.Check(err, gc.ErrorMatches, expectErr)
		return nil
	}

	c.Assert(err, jc.ErrorIsNil)
	c.Assert(p, gc.NotNil)
	c.Assert(p.C, gc.NotNil)
	// should not be closed yet
	c.Assert(qt.IsReplyClosed(*reply), gc.Equals, false)

	testSend(c, p, messageQueue, messages)
	return p
}

func testSend(c *gc.C, ch *queue.PubChannel, messageQueue *[][]byte, msgs [][]byte) {
	c.Assert(ch, gc.NotNil)

	// send some messages
	err := qt.TrySend(ch, msgs)
	c.Assert(err, jc.ErrorIsNil)

	qt.SyncGC()
	c.Check(*messageQueue, jc.DeepEquals, msgs)
}
