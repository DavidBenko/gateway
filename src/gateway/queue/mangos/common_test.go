package mangos_test

import (
	"reflect"
	"time"

	"gateway/queue"
	qm "gateway/queue/mangos"
	"gateway/queue/testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func getBasicPub(c *gc.C, path string) queue.Publisher {
	p, err := queue.Publish(
		path,
		qm.Pub,
		qm.PubTCP,
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(p, gc.NotNil)
	c.Assert(reflect.TypeOf(p), gc.Equals, reflect.TypeOf(&qm.PubSocket{}))

	return p
}

func getBasicSub(c *gc.C, path string) queue.Subscriber {
	s, err := queue.Subscribe(
		path,
		qm.Sub,
		qm.SubTCP,
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(s, gc.NotNil)
	c.Assert(reflect.TypeOf(s), gc.Equals, reflect.TypeOf(&qm.SubSocket{}))

	return s
}

func testPubSub(c *gc.C, pub queue.Publisher, sub queue.Subscriber, msg string, shouldReceive bool) {
	pCh := pub.Channel()
	sCh := sub.Channel()

	c.Check(pCh, gc.NotNil)
	c.Check(sCh, gc.NotNil)

	go func() { pCh <- []byte(msg) }()

	select {
	case received := <-sCh:
		switch shouldReceive {
		case true:
			c.Check(string(received), gc.Equals, msg)
		case false:
			c.Logf("testPubSub: Received unintended message %q", received)
			c.FailNow()
		}
	case <-time.After(testing.ShortWait):
		switch shouldReceive {
		case true:
			c.Logf("testPubSub: Never received message")
			c.FailNow()
		case false:
		}
	}
}
