package mangos_test

import (
	"fmt"
	"gateway/queue/mangos"
	"runtime"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func (s *MangosSuite) TestNewBroker(c *gc.C) {
	for i, t := range []struct {
		should      string
		k           mangos.Kind
		transport   mangos.Transport
		xPubPath    string
		xSubPath    string
		expectError string
	}{{
		should:      "fail with unknown Kind",
		k:           mangos.Kind(-1),
		transport:   mangos.TCP,
		expectError: "unknown Broker Kind -1",
	}, {
		should:      "fail with unknown Transport",
		k:           mangos.XPubXSub,
		transport:   mangos.Transport(-1),
		expectError: "unknown Broker Transport -1",
	}, {
		should:    "make a working TCP Broker",
		k:         mangos.XPubXSub,
		transport: mangos.TCP,
		xPubPath:  "tcp://localhost:9000",
		xSubPath:  "tcp://localhost:9001",
	}, {
		should:    "make a working IPC Broker (but only on Linux or Darwin)",
		k:         mangos.XPubXSub,
		transport: mangos.IPC,
		xPubPath:  brokerIPCPub,
		xSubPath:  brokerIPCSub,
	}} {
		c.Logf("test %d: should %s", i, t.should)

		b, err := mangos.NewBroker(
			t.k,
			t.transport,
			t.xPubPath, t.xSubPath,
		)

		if t.transport == mangos.IPC && !ipcSupported() {
			c.Assert(err, gc.ErrorMatches, fmt.Sprintf(
				".* failed: mangos IPC transport not supported on OS %q", runtime.GOOS))
			continue
		} else if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			continue
		}

		c.Assert(err, jc.ErrorIsNil)

		p := getBrokeredPub(c, t.xSubPath, t.transport)
		s := getBasicSub(c, t.xPubPath, t.transport)

		c.Assert(p, gc.NotNil)
		c.Assert(s, gc.NotNil)

		testPubSub(c, p, s, "hello", true)

		c.Assert(b.Close(), jc.ErrorIsNil)
	}
}
