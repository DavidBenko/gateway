package mangos_test

import (
	"os"
	"reflect"
	"runtime"
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

func getBasicPub(c *gc.C, path string, trans ...qm.Transport) queue.Publisher {
	options := []queue.PubBinding{
		qm.Pub(false),
		qm.PubBuffer(2048),
	}

	if len(trans) == 0 {
		options = append(options, qm.PubTCP)
	}

	for _, tran := range trans {
		switch tran {
		case qm.TCP:
			options = append(options, qm.PubTCP)
		case qm.IPC:
			options = append(options, qm.PubIPC)
		default:
			c.Logf("getBasicPub does not support transport %d", tran)
			c.FailNow()
		}
	}

	p, err := queue.Publish(
		path,
		options...,
	)

	c.Assert(err, jc.ErrorIsNil)
	c.Assert(p, gc.NotNil)
	c.Assert(reflect.TypeOf(p), gc.Equals, reflect.TypeOf(&qm.PubSocket{}))

	return p
}

func getBrokeredPub(c *gc.C, path string, trans ...qm.Transport) queue.Publisher {
	options := []queue.PubBinding{
		qm.Pub(true),
		qm.PubBuffer(2048),
	}

	if len(trans) == 0 {
		options = append(options, qm.PubTCP)
	}

	for _, tran := range trans {
		switch tran {
		case qm.TCP:
			options = append(options, qm.PubTCP)
		case qm.IPC:
			options = append(options, qm.PubIPC)
		default:
			c.Logf("getBrokeredPub does not support transport %d", tran)
			c.FailNow()
		}
	}

	p, err := queue.Publish(
		path,
		options...,
	)

	c.Assert(err, jc.ErrorIsNil)
	c.Assert(p, gc.NotNil)
	c.Assert(reflect.TypeOf(p), gc.Equals, reflect.TypeOf(&qm.PubSocket{}))

	c.Assert(qm.IsBrokered(c, p), gc.Equals, true)

	return p
}

func getBasicSub(c *gc.C, path string, trans ...qm.Transport) queue.Subscriber {
	options := []queue.SubBinding{
		qm.Sub,
		qm.SubBuffer(2048),
	}

	if len(trans) == 0 {
		options = append(options, qm.SubTCP)
	}

	for _, tran := range trans {
		switch tran {
		case qm.TCP:
			options = append(options, qm.SubTCP)
		case qm.IPC:
			options = append(options, qm.SubIPC)
		default:
			c.Logf("getBasicSub does not support transport %d", tran)
			c.FailNow()
		}
	}

	s, err := queue.Subscribe(
		path,
		options...,
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
	// PubSub does not guarantee 100 percent delivery.
	acceptableRate := rate > 0.9
	c.Logf("testPubSub: Received %d messages out of %d", total, TotalAttempts)
	c.Logf("testPubSub:   --- %f success rate ---", rate)
	c.Check(acceptableRate, gc.Equals, true)
}

func trySend(
	c *gc.C,
	msg string,
	pCh chan<- []byte,
	pE <-chan error,
	doneSend chan struct{},
) {
	to := time.After(testing.LongWait)
	for i := 0; i < TotalAttempts; i++ {
		select {
		case e := <-pE:
			c.Logf("unexpected receive from error channel: %#v", e)
			c.FailNow()
		case pCh <- []byte(msg):
		case <-to:
			c.Logf("testPubSub: failed to send after %s",
				testing.LongWait.String())
			c.FailNow()
		}
		time.Sleep(10000)
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
		// Finished without receiving anything, which is the desired
		// behavior.
	case <-time.After(testing.LongWait):
		c.Logf("testPubSub: tryShouldNotReceive timed out after %s",
			testing.LongWait.String())
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
		case e := <-sE:
			c.Logf("unexpected receive from error channel: %#v", e)
			c.FailNow()
		case m := <-sCh:
			c.Check(string(m), gc.Equals, msg)
			total++
		case <-time.After(testing.LongWait):
			break Recv
		}
	}
	doneRecv <- total
}

func ipcSupported() bool {
	switch runtime.GOOS {
	case "darwin", "linux":
		return true
	default:
		return false
	}
}

func clearIPCFiles() error {
	for _, ipc := range ipcFiles() {
		_, err := os.Stat(ipc)
		switch {
		case err != nil && os.IsNotExist(err):
			return nil
		case err != nil:
			return err
		default:
			if err = os.Remove(ipc); err != nil {
				return err
			}
		}
	}

	return nil
}
