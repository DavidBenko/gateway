package conversion_test

import (
	"gateway/core/vm/conversion"
	"testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func TestPath(t *testing.T) { gc.TestingT(t) }

type PathSuite struct{}

var _ = gc.Suite(&PathSuite{})

func (s *PathSuite) TestXmlPath(c *gc.C) {
	stub := "<doc><meta><id>111</id><name>foobar</name></meta></doc>"
	var sub []string

	result, err := conversion.XMLPath(stub, "doc.meta.id", sub)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(len(result), gc.Equals, 1)

	id := result[0].(string)
	c.Assert(id, gc.Equals, "111")
}

func (s *PathSuite) TestXmlPathInvalidXml(c *gc.C) {
	stub := "<doc><meta><id>111</id><name>foobar</name></meta>"
	var sub []string

	_, err := conversion.XMLPath(stub, "doc.meta.id", sub)
	c.Assert(err, gc.NotNil)
}

func (s *PathSuite) TestingXmlPathInvalidPath(c *gc.C) {
	stub := "<doc><meta><id>111</id><name>foobar</name></meta>"
	var sub []string

	result, err := conversion.XMLPath(stub, "user.id", sub)
	c.Assert(err, gc.NotNil)
	c.Assert(len(result), gc.Equals, 0)
}

func (s *PathSuite) TestJsonPath(c *gc.C) {
	json := map[string]interface{}{
		"meta": map[string]interface{}{
			"id":   "111",
			"name": "foobar",
		},
	}

	var sub []string
	result, err := conversion.JSONPath(json, "meta.id", sub)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(len(result), gc.Equals, 1)

	id := result[0].(string)
	c.Assert(id, gc.Equals, "111")
}

func (s *PathSuite) TestJsonPathInvalidPath(c *gc.C) {
	json := map[string]interface{}{
		"meta": map[string]interface{}{
			"id":   "111",
			"name": "foobar",
		},
	}

	var sub []string
	result, err := conversion.JSONPath(json, "user.id", sub)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(len(result), gc.Equals, 0)
}
