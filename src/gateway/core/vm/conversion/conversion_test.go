package conversion_test

import (
	"testing"

	"gateway/core/vm/conversion"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func TestConversion(t *testing.T) { gc.TestingT(t) }

type ConversionSuite struct{}

var _ = gc.Suite(&ConversionSuite{})

func (s *ConversionSuite) TestJsonToXml(c *gc.C) {
	stub := map[string]interface{}{
		"foo": "bar",
		"nested": map[string]interface{}{
			"key": "value",
		},
	}

	result, err := conversion.ToXML(stub)

	c.Assert(err, jc.ErrorIsNil)

	c.Assert(result, gc.Equals, "<doc><foo>bar</foo><nested><key>value</key></nested></doc>")
}

func (s *ConversionSuite) TestXmlToJson(c *gc.C) {
	stub := "<doc><foo>bar</foo><nested><key>value</key></nested></doc>"

	result, err := conversion.ToJSON(stub)

	c.Assert(err, jc.ErrorIsNil)

	c.Assert(result, gc.Equals, "{\"doc\":{\"foo\":\"bar\",\"nested\":{\"key\":\"value\"}}}")
}
