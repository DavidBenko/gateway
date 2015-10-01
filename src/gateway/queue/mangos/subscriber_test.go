package mangos_test

import (
	"fmt"
	"gateway/queue"
	qm "gateway/queue/mangos"
	"gateway/queue/testing"
	"reflect"
	"runtime"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func (s *MangosSuite) TestSubSocket(c *gc.C) {
	m := &qm.SubSocket{}
	err := m.Connect("foo")
	c.Logf("TestSubSocket: SubSocket can't Bind on nil socket")
	c.Check(err, gc.ErrorMatches, "mangos Subscriber couldn't Connect to foo: nil socket")

	err = m.Close() // Does nothing
	c.Logf("TestSubSocket: SubSocket Close with nil socket does nothing")
	c.Check(err, jc.ErrorIsNil)

	c.Logf("TestSubSocket: pub-sub works correctly")
	pub := getBasicPub(c, "tcp://localhost:9001")
	sub := getBasicSub(c, "tcp://localhost:9001")

	testPubSub(c, pub, sub, "hello", true)

	c.Logf("TestSubSocket: live SubSocket Close does not error")
	c.Assert(sub.Close(), jc.ErrorIsNil)
	c.Logf("TestSubSocket: live PubSocket Close does not error")
	c.Assert(pub.Close(), jc.ErrorIsNil)
}

func (s *MangosSuite) TestGetSubSocket(c *gc.C) {
	sc, err := qm.GetSubSocket(&qm.SubSocket{})
	c.Check(err, jc.ErrorIsNil)
	c.Check(sc, gc.IsNil)

	sc, err = qm.GetSubSocket(&testing.Subscriber{})
	c.Check(err, gc.ErrorMatches, `getSubSocket expected \*mangos.SubSocket, got \*testing.Subscriber`)

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
	c.Check(err, gc.ErrorMatches, "SubTCP requires a non-nil Socket, use Sub first")
}

func (s *MangosSuite) TestSubIPC(c *gc.C) {
	switch runtime.GOOS {
	case "linux", "darwin":
	default:
		c.Logf(`TestSubIPC: supported only on runtime.GOOS == "linux" or "darwin"`)
		_, err := queue.Subscribe(
			"ipc:///tmp/test.ipc",
			qm.Sub,
			qm.SubIPC,
		)
		c.Check(err, gc.ErrorMatches, fmt.Sprintf("SubIPC failed: mangos IPC transport not supported on OS %q", runtime.GOOS))
		return // Don't need to test other behaviors
	}

	c.Logf("TestSubIPC: nil socket fails")
	_, err := qm.SubIPC(&qm.SubSocket{})
	c.Check(err, gc.ErrorMatches, "SubIPC requires a non-nil Socket, use Sub first")

	c.Logf("TestSubIPC: pub-sub works on IPC")
	pIPC, err := queue.Publish(
		"ipc:///tmp/ipc.ipc",
		qm.Pub,
		qm.PubIPC,
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(pIPC, gc.NotNil)

	c.Logf("TestSubIPC: correct usage works")
	sIPC, err := queue.Subscribe(
		"ipc:///tmp/ipc.ipc",
		qm.Sub,
		qm.SubIPC,
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(sIPC, gc.NotNil)

	testPubSub(c, pIPC, sIPC, "hello", true)

	c.Logf("TestSubIPC: live SubSocket Close does not error")
	c.Assert(sIPC.Close(), jc.ErrorIsNil)
	c.Logf("TestSubIPC: live PubSocket Close does not error")
	c.Assert(pIPC.Close(), jc.ErrorIsNil)
}

func (s *MangosSuite) TestFilter(c *gc.C) {
	var ss *qm.SubSocket
	sub, err := qm.Filter("")(ss) //queue.Subscriber(nil))
	c.Assert(err, gc.ErrorMatches, `Filter got nil Subscriber, use Pub first`)

	sub, err = qm.Filter("")(queue.Subscriber(nil))
	c.Assert(err, gc.ErrorMatches, `Filter got nil Subscriber, use Pub first`)

	sub, err = qm.Filter("")(&testing.Subscriber{})
	c.Assert(err, gc.ErrorMatches, `Filter expected \*mangos.SubSocket, got \*testing.Subscriber`)

	pub := getBasicPub(c, "tcp://localhost:9001")

	sub, err = queue.Subscribe(
		"tcp://localhost:9001",
		qm.Sub,
		qm.SubTCP,
		qm.Filter("foo"),
	)

	testPubSub(c, pub, sub, "foo|hello", true)
	testPubSub(c, pub, sub, "hello", false)
	c.Logf("TestFilter: live SubSocket Close does not error")
	c.Assert(sub.Close(), jc.ErrorIsNil)
	c.Logf("TestFilter: live PubSocket Close does not error")
	c.Assert(pub.Close(), jc.ErrorIsNil)
}

func (s *MangosSuite) TestSub(c *gc.C) {
	sub, err := qm.Sub(&testing.Subscriber{})
	c.Assert(err, gc.ErrorMatches, `mangos.Sub expects nil Subscriber, got \*testing.Subscriber`)

	var qs queue.Subscriber
	sub, err = qm.Sub(queue.Subscriber(qs))
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(sub, gc.NotNil)
	c.Assert(reflect.TypeOf(sub), gc.Equals, reflect.TypeOf(&qm.SubSocket{}))
	c.Assert(sub.Close(), jc.ErrorIsNil)
}
