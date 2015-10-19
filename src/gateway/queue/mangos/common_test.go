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

const (
	// TotalAttempts specifies how many messages to send on testPubSub.
	TotalAttempts = 1000
)

func getBasicPub(c *gc.C, path string) queue.Publisher {
	p, err := queue.Publish(
		path,
		qm.Pub,
		qm.PubTCP,
		qm.PubBuffer(2048),
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
		qm.SubBuffer(2048),
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(s, gc.NotNil)
	c.Assert(reflect.TypeOf(s), gc.Equals, reflect.TypeOf(&qm.SubSocket{}))

	return s
}

func testPubSub(c *gc.C, pub queue.Publisher, sub queue.Subscriber, msg string, shouldReceive bool) {
	pCh, pE := pub.Channels()
	sCh, sE := sub.Channels()

	c.Check(pCh, gc.NotNil)
	c.Check(sCh, gc.NotNil)

	doneSend := make(chan struct{})
	doneRecv := make(chan int)

	// Try some sends
	go trySend(c, msg, pCh, pE, doneSend)
	if !shouldReceive {
		tryShouldNotReceive(c, msg, sCh, sE, doneSend)
		return
	}

	// Otherwise, make sure the received count adds up.
	go tryShouldReceive(c, msg, sCh, sE, doneRecv)

	<-doneSend
	total := <-doneRecv

	rate := float64(total) / float64(TotalAttempts)
	c.Logf("testPubSub: Received %d messages out of %d", total, TotalAttempts)
	c.Logf("testPubSub:   --- %f success rate ---", rate)
	c.Check(rate, gc.Equals, 1.0)
}

func trySend(
	c *gc.C,
	msg string,
	pCh chan<- []byte,
	pE <-chan error,
	doneSend chan struct{},
) {
	to := time.After(5 * time.Second)
	for i := 0; i < TotalAttempts; i++ {
		select {
		case e, ok := <-pE:
			if !ok {
				c.Log("testPubSub: error channel was closed")
				c.FailNow()
			}
			c.Assert(e, jc.ErrorIsNil)
			i--
		case pCh <- []byte(msg):
		case <-to:
			c.Log("testPubSub: failed to send after 5 seconds")
			c.FailNow()
		}
	}
	close(doneSend)
}

func tryShouldNotReceive(
	c *gc.C,
	msg string,
	sCh <-chan []byte,
	sE <-chan error,
	doneSend chan struct{},
) {
Recv:
	select {
	case e, ok := <-sE:
		c.Assert(e, jc.ErrorIsNil)
		if ok {
			c.Logf("testPubSub: Received unexpected nil error %v", e)
			goto Recv
		}
		c.Logf("testPubSub: Received unexpected nil error %q", msg)
		c.FailNow()
	case msg := <-sCh:
		c.Logf("testPubSub: Received unintended message %q", msg)
		c.FailNow()
	case <-doneSend:
		c.Log("testPubSub: Received no messages, as intended")
		// Finished without receiving anything, which is the
		// desired behavior.
	case <-time.After(5 * time.Second):
		c.Log("testPubSub: tryShouldNotReceive timed out after 5 seconds")
		c.FailNow()
	}
	return
}

func tryShouldReceive(
	c *gc.C,
	msg string,
	sCh <-chan []byte,
	sE <-chan error,
	doneRecv chan int,
) {
	total := 0
Recv:
	for {
		select {
		case e, ok := <-sE:
			if !ok {
				c.Log("testPubSub: error channel was closed")
				c.FailNow()
			}
			c.Assert(e, jc.ErrorIsNil)
		case m := <-sCh:
			c.Check(string(m), gc.Equals, msg)
			total++
		case <-time.After(testing.LongWait):
			break Recv
		}
	}
	doneRecv <- total
}
