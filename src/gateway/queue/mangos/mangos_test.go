package mangos_test

import (
	"testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

const (
	brokerIPCPub = "ipc:///tmp/broker-test-pub.ipc"
	brokerIPCSub = "ipc:///tmp/broker-test-sub.ipc"

	ipcTest = "ipc:///tmp/test.ipc"
)

func ipcFiles() []string {
	return []string{
		brokerIPCPub,
		brokerIPCSub,
		ipcTest,
	}
}

func Test(t *testing.T) { gc.TestingT(t) }

type MangosSuite struct{}

var _ = gc.Suite(&MangosSuite{})

func (s *MangosSuite) SetUpTest(c *gc.C) {
	c.Assert(clearIPCFiles(), jc.ErrorIsNil)
}

func (s *MangosSuite) TearDownTest(c *gc.C) {
	c.Assert(clearIPCFiles(), jc.ErrorIsNil)
}
