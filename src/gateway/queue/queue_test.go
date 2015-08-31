package queue_test

import (
	"gateway/queue"
	"gateway/queue/testing"

	gc "gopkg.in/check.v1"

	jc "github.com/juju/testing/checkers"
)

type QueueSuite struct{}

var _ = gc.Suite(&QueueSuite{})

func (s *QueueSuite) TestTestingQueue(c *gc.C) {
	p, err := queue.PubChannel(
		"local",
		testing.Publish(),
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(p, gc.NotNil)
}
