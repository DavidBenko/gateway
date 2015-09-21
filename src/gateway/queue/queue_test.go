package queue_test

import (
	"testing"

	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) { gc.TestingT(t) }

type QueueSuite struct{}

var _ = gc.Suite(&QueueSuite{})
