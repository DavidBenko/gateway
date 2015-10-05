package mangos_test

import (
	"testing"

	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) { gc.TestingT(t) }

type MangosSuite struct{}

var _ = gc.Suite(&MangosSuite{})
