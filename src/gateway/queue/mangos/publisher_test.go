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

func (s *MangosSuite) TestPubSocket(c *gc.C) {
	m := &qm.PubSocket{}
	err := m.Bind("foo")
	c.Logf("PubSocket can't Bind on nil socket")
	c.Check(err, gc.ErrorMatches, "mangos Publisher couldn't Bind to foo: nil socket")

	err = m.Close() // Does nothing
	c.Logf("PubSocket Close with nil socket does nothing")
	c.Check(err, jc.ErrorIsNil)

	p := getBasicPub(c, "tcp://localhost:9001")

	ch := p.Channel()
	c.Check(ch, gc.NotNil)

	c.Logf("live PubSocket Close does not error")
	c.Assert(p.Close(), jc.ErrorIsNil)
}

func (s *MangosSuite) TestGetPubSocket(c *gc.C) {
	sc, err := qm.GetPubSocket(&qm.PubSocket{})
	c.Check(err, jc.ErrorIsNil)
	c.Check(sc, gc.IsNil)

	sc, err = qm.GetPubSocket(&testing.Publisher{})
	c.Check(err, gc.ErrorMatches, `GetPubSocket expected \*mangos.PubSocket, got \*testing.Publisher`)

	p := getBasicPub(c, "tcp://localhost:9001")
	defer c.Assert(p.Close(), jc.ErrorIsNil)

	sc, err = qm.GetPubSocket(p)

	c.Assert(err, jc.ErrorIsNil)
	c.Assert(sc, gc.NotNil)
}

func (s *MangosSuite) TestPubTCP(c *gc.C) {
	pIPC, err := queue.Publish(
		"tcp://localhost:9001",
		qm.Pub,
		qm.PubTCP,
	)

	c.Assert(err, jc.ErrorIsNil)
	c.Assert(pIPC, gc.NotNil)
	err = pIPC.Close()

	c.Assert(err, jc.ErrorIsNil)

	_, err = qm.PubIPC(&qm.PubSocket{})
	c.Check(err, gc.ErrorMatches, "PubIPC requires a non-nil Socket, use Pub first")
}

func (s *MangosSuite) TestPubIPC(c *gc.C) {
	pIPC, err := queue.Publish(
		"ipc:///tmp/test.ipc",
		qm.Pub,
		qm.PubIPC,
	)

	switch runtime.GOOS {
	case "linux", "darwin":
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(pIPC, gc.NotNil)
		err = pIPC.Close()
		c.Assert(err, jc.ErrorIsNil)
	default:
		c.Check(err, gc.ErrorMatches, fmt.Sprintf("PubIPC failed: mangos IPC transport not supported on OS %q", runtime.GOOS))
		return // Don't need to test other behaviors
	}

	_, err = qm.PubIPC(&qm.PubSocket{})
	c.Check(err, gc.ErrorMatches, "PubIPC requires a non-nil Socket, use Pub first")
}

func (s *MangosSuite) TestPub(c *gc.C) {
	p, err := qm.Pub(&testing.Publisher{})
	c.Assert(err, gc.ErrorMatches, `mangos.Pub expects nil Publisher, got \*testing.Publisher`)

	var qp queue.Publisher
	p, err = qm.Pub(queue.Publisher(qp))
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(p, gc.NotNil)
	c.Assert(reflect.TypeOf(p), gc.Equals, reflect.TypeOf(&qm.PubSocket{}))
	qP := p.(*qm.PubSocket)
	sock := qP.Socket()
	c.Assert(sock, gc.NotNil)
	c.Assert(p.Close(), jc.ErrorIsNil)
}
