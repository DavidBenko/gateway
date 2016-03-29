package model_test

import (
	"testing"

	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) { gc.TestingT(t) }

type ModelSuite struct{}

var _ = gc.Suite(&ModelSuite{})
