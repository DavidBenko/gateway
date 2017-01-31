package vm_test

import (
	"gateway/core"
	"gateway/core/vm"
	"testing"

	jc "github.com/juju/testing/checkers"

	gc "gopkg.in/check.v1"
)

func TestVM(t *testing.T) { gc.TestingT(t) }

type VMSuite struct{}

var _ = gc.Suite(&VMSuite{})

func (s *VMSuite) TestExtensionsIncluded(c *gc.C) {
	k := &vm.KeyStore{}
	e := &vm.RemoteEndpointStore{}

	vm := core.VMCopy(1, k, e, nil)

	for i, t := range []struct {
		should      string
		get         string
		expectClass string
	}{{
		should:      "include AP.Crypto",
		get:         "AP.Crypto",
		expectClass: "Object",
	}, {
		should:      "include AP.Crypto.encrypt",
		get:         "AP.Crypto.encrypt",
		expectClass: "Function",
	}, {
		should:      "include AP.Crypto.decrypt",
		get:         "AP.Crypto.decrypt",
		expectClass: "Function",
	}, {
		should:      "include AP.Crypto.hashPassword",
		get:         "AP.Crypto.hashPassword",
		expectClass: "Function",
	}, {
		should:      "include AP.Crypto.compareHashAndPassword",
		get:         "AP.Crypto.compareHashAndPassword",
		expectClass: "Function",
	}, {
		should:      "include AP.Crypto.hash",
		get:         "AP.Crypto.hash",
		expectClass: "Function",
	}, {
		should:      "include AP.Crypto.hashHmac",
		get:         "AP.Crypto.hashHmac",
		expectClass: "Function",
	}, {
		should:      "include AP.Crypto.sign",
		get:         "AP.Crypto.sign",
		expectClass: "Function",
	}, {
		should:      "include AP.Crypto.verify",
		get:         "AP.Crypto.verify",
		expectClass: "Function",
	}, {
		should:      "include AP.Crypto.Aes",
		get:         "AP.Crypto.Aes",
		expectClass: "Object",
	}, {
		should:      "include AP.Crypto.Aes.encrypt",
		get:         "AP.Crypto.Aes.encrypt",
		expectClass: "Function",
	}, {
		should:      "include AP.Crypto.Aes.decrypt",
		get:         "AP.Crypto.Aes.decrypt",
		expectClass: "Function",
	}, {
		should:      "include AP.Encoding",
		get:         "AP.Encoding",
		expectClass: "Object",
	}, {
		should:      "include AP.Encoding.toBase64",
		get:         "AP.Encoding.toBase64",
		expectClass: "Function",
	}, {
		should:      "include AP.Encoding.fromBase64",
		get:         "AP.Encoding.fromBase64",
		expectClass: "Function",
	}, {
		should:      "include AP.Encoding.toHex",
		get:         "AP.Encoding.toHex",
		expectClass: "Function",
	}, {
		should:      "include AP.Encoding.fromHex",
		get:         "AP.Encoding.fromHex",
		expectClass: "Function",
	}, {
		should:      "include AP.Conversion",
		get:         "AP.Conversion",
		expectClass: "Object",
	}, {
		should:      "include AP.Conversion.toJson",
		get:         "AP.Conversion.toJson",
		expectClass: "Function",
	}, {
		should:      "include AP.Conversion.JSONPath",
		get:         "AP.Conversion.JSONPath",
		expectClass: "Function",
	}, {
		should:      "include AP.Conversion.toXML",
		get:         "AP.Conversion.toXML",
		expectClass: "Function",
	}, {
		should:      "include AP.Conversion.XMLPath",
		get:         "AP.Conversion.XMLPath",
		expectClass: "Function",
	}, {
		should:      "include AP.Perform",
		get:         "AP.Perform",
		expectClass: "Function",
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		v, err := vm.Object(t.get)
		c.Assert(err, jc.ErrorIsNil)

		if v.Class() != t.expectClass {
			c.Errorf("expected type %s got type %s", t.expectClass, v.Class())
		}
	}
}
