package mangos_test

import (
	"fmt"
	"gateway/queue"
	qm "gateway/queue/mangos"
	"gateway/queue/testing"
	"reflect"
	"runtime"

	"github.com/go-mangos/mangos"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func (s *MangosSuite) TestSubSocket(c *gc.C) {
	m := &qm.SubSocket{}
	err := m.Connect("foo")
	c.Log("TestSubSocket: SubSocket can't Bind on nil socket")
	c.Check(err, gc.ErrorMatches, "SubSocket couldn't Connect to foo: nil socket")

	err = m.Close() // Does nothing
	c.Log("TestSubSocket: SubSocket Close with nil socket does nothing")
	c.Check(err, jc.ErrorIsNil)

	c.Log("TestSubSocket: pub-sub works correctly")
	pub := getBasicPub(c, "tcp://localhost:9001")
	sub := getBasicSub(c, "tcp://localhost:9001")

	testPubSub(c, pub, sub, "hello", true)

	c.Log("TestSubSocket: live SubSocket Close does not error")
	c.Assert(sub.Close(), jc.ErrorIsNil)
	c.Log("TestSubSocket: live PubSocket Close does not error")
	c.Assert(pub.Close(), jc.ErrorIsNil)
}

func (s *MangosSuite) TestGetSubSocket(c *gc.C) {
	sc, err := qm.GetSubSocket(&qm.SubSocket{})
	c.Check(err, jc.ErrorIsNil)
	c.Check(sc, gc.IsNil)

	sc, err = qm.GetSubSocket(&testing.Subscriber{})
	c.Check(err, gc.ErrorMatches, `getSubSocket expected \*SubSocket, got \*testing.Subscriber`)

	p := getBasicSub(c, "tcp://localhost:9001")

	sc, err = qm.GetSubSocket(p)

	c.Assert(err, jc.ErrorIsNil)
	c.Assert(sc, gc.NotNil)
	c.Assert(p.Close(), jc.ErrorIsNil)
}

func (s *MangosSuite) TestSubTCP(c *gc.C) {
	sTCP, err := queue.Subscribe(
		"tcp://localhost:9001",
		qm.Sub,
		qm.SubTCP,
	)

	c.Assert(err, jc.ErrorIsNil)
	c.Assert(sTCP, gc.NotNil)
	err = sTCP.Close()

	c.Assert(err, jc.ErrorIsNil)

	_, err = qm.SubTCP(&qm.SubSocket{})
	c.Check(err, gc.ErrorMatches, "SubTCP requires a non-nil Socket, use Sub or XSub first")
}

func (s *MangosSuite) TestSubIPC(c *gc.C) {
	switch runtime.GOOS {
	case "linux", "darwin":
	default:
		c.Log(`TestSubIPC: supported only on runtime.GOOS == "linux" or "darwin"`)
		_, err := queue.Subscribe(
			ipcTest,
			qm.Sub,
			qm.SubIPC,
		)
		c.Check(err, gc.ErrorMatches, fmt.Sprintf("SubIPC failed: IPC transport not supported on OS %q", runtime.GOOS))
		return // Don't need to test other behaviors
	}

	c.Log("TestSubIPC: nil socket fails")
	_, err := qm.SubIPC(&qm.SubSocket{})
	c.Check(err, gc.ErrorMatches, "SubIPC requires a non-nil Socket, use Sub or XSub first")

	c.Log("TestSubIPC: pub-sub works on IPC")
	pIPC, err := queue.Publish(
		ipcTest,
		qm.Pub(false),
		qm.PubIPC,
		qm.PubBuffer(2048),
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(pIPC, gc.NotNil)

	c.Log("TestSubIPC: correct usage works")
	sIPC, err := queue.Subscribe(
		ipcTest,
		qm.Sub,
		qm.SubIPC,
		qm.SubBuffer(2048),
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(sIPC, gc.NotNil)

	testPubSub(c, pIPC, sIPC, "hello", true)

	c.Log("TestSubIPC: live SubSocket Close does not error")
	c.Assert(sIPC.Close(), jc.ErrorIsNil)
	c.Log("TestSubIPC: live PubSocket Close does not error")
	c.Assert(pIPC.Close(), jc.ErrorIsNil)
}

func (s *MangosSuite) TestFilter(c *gc.C) {
	var ss *qm.SubSocket
	sub, err := qm.Filter("")(ss) //queue.Subscriber(nil))
	c.Assert(err, gc.ErrorMatches, `Filter got nil Subscriber, use Pub first`)

	sub, err = qm.Filter("")(queue.Subscriber(nil))
	c.Assert(err, gc.ErrorMatches, `Filter got nil Subscriber, use Pub first`)

	sub, err = qm.Filter("")(&testing.Subscriber{})
	c.Assert(err, gc.ErrorMatches, `Filter expected \*SubSocket, got \*testing.Subscriber`)

	pub := getBasicPub(c, "tcp://localhost:9001")

	sub, err = queue.Subscribe(
		"tcp://localhost:9001",
		qm.Sub,
		qm.SubTCP,
		qm.SubBuffer(2048),
		qm.Filter("foo"),
	)

	testPubSub(c, pub, sub, "foo|hello", true)
	testPubSub(c, pub, sub, "hello", false)
	c.Log("TestFilter: live SubSocket Close does not error")
	c.Assert(sub.Close(), jc.ErrorIsNil)
	c.Log("TestFilter: live PubSocket Close does not error")
	c.Assert(pub.Close(), jc.ErrorIsNil)
}

func (s *MangosSuite) TestSub(c *gc.C) {
	sub, err := qm.Sub(&testing.Subscriber{})
	c.Assert(err, gc.ErrorMatches, `Sub expects nil Subscriber, got \*testing.Subscriber`)

	sub, err = qm.Sub(queue.Subscriber(nil))
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(sub, gc.NotNil)
	c.Assert(reflect.TypeOf(sub), gc.Equals, reflect.TypeOf(&qm.SubSocket{}))
	c.Assert(sub.Close(), jc.ErrorIsNil)
}

func (s *MangosSuite) TestXSub(c *gc.C) {
	sub, err := qm.XSub(&testing.Subscriber{})
	c.Assert(err, gc.ErrorMatches, `XSub expects nil Subscriber, got \*testing.Subscriber`)

	sub, err = qm.XSub(queue.Subscriber(nil))
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(sub, gc.NotNil)
	c.Assert(reflect.TypeOf(sub), gc.Equals, reflect.TypeOf(&qm.SubSocket{}))

	xsSock, err := qm.GetSubSocket(sub)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(xsSock, gc.NotNil)
	isRawIf, err := xsSock.GetOption(mangos.OptionRaw)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(
		reflect.TypeOf(isRawIf),
		gc.Equals,
		reflect.TypeOf(interface{}(true)),
	)
	isRaw := isRawIf.(bool)
	c.Check(isRaw, gc.Equals, true)

	c.Assert(sub.Close(), jc.ErrorIsNil)
}

func (s *MangosSuite) TestSubBufferSize(c *gc.C) {
	_, err := qm.SubBuffer(-10)(&qm.SubSocket{})
	c.Assert(err, gc.ErrorMatches, "SubBuffer expects positive size, got -10")

	_, err = qm.SubBuffer(10)(&testing.Subscriber{})
	c.Assert(err, gc.ErrorMatches, `getSubSocket expected \*SubSocket, got \*testing.Subscriber`)

	ss, err := qm.Sub(queue.Subscriber(nil))
	c.Assert(err, jc.ErrorIsNil)
	subs, err := qm.SubBuffer(10)(ss)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(qm.GetSubBufferSize(subs), gc.Equals, 10)

	subs, err = queue.Subscribe(
		"tcp://localhost:9001",
		qm.Sub,
		qm.SubTCP,
		qm.SubBuffer(2048),
	)
	c.Assert(err, jc.ErrorIsNil)

	// Make sure it was set correctly on the socket itself
	sock, err := qm.GetSubSocket(subs)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(subs, gc.NotNil)

	buffSize, err := sock.GetOption(mangos.OptionReadQLen)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(buffSize, gc.Equals, 2048)

	c.Assert(subs.Close(), jc.ErrorIsNil)
}
