package util_test

import (
	"fmt"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

type UtilSuite struct{}

var _ = gc.Suite(&UtilSuite{})

func (s *UtilSuite) TestDrain(c *gc.C) {
	c.Assert(fmt.Errorf("foo"), jc.ErrorIsNil)
}
