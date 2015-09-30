package mangos_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) { gc.TestingT(t) }

type MangosSuite struct {
	port int
	ipc  []int
}

var _ = gc.Suite(&MangosSuite{})

func (m *MangosSuite) SetUpSuite(c *gc.C) {
	m.port = 9000
	m.ipc = []int{0}
}

func (m *MangosSuite) TCPURI(c *gc.C) string {
	t := time.After(3 * time.Second)
	for p := m.port; portOccupied(p, c); p++ {
		if p > 65535 {
			p = 9000
		}
		select {
		case <-t:
			c.Log("test failed to find a free port")
			c.FailNow()
			return ""
		default:
			m.port = p
		}
	}

	return fmt.Sprintf("tcp://localhost:%d", m.port)
}

func (m *MangosSuite) IPCURI(c *gc.C) string {
	t := time.After(3 * time.Second)
	for p := m.port; portOccupied(p, c); p++ {
		if p > 65535 {
			p = 9000
		}
		select {
		case <-t:
			c.Log("test failed to find a free port")
			c.FailNow()
			return ""
		default:
			m.port = p
		}
	}

	return fmt.Sprintf("tcp://localhost:%d", m.port)
}

// True if port is free, false otherwise
func portOccupied(p int, c *gc.C) bool {
	addrStr := fmt.Sprintf("localhost:%d", p)
	addr, err := net.ResolveTCPAddr("tcp", addrStr)
	if err != nil {
		c.Logf("bad TCP socket %q: %v", addrStr, err)
		c.FailNow()
	}

	t, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return true
	}

	if e := t.Close(); e != nil {
		c.Logf("verifyPort failed to close socket %d: %v", p, e)
		c.FailNow()
	}
	return false
}
