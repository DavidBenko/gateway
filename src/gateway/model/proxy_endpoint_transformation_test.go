package model_test

import (
	aperrors "gateway/errors"
	"gateway/model"
	"strconv"

	jc "github.com/juju/testing/checkers"
	"github.com/robertkrimen/otto"
	gc "gopkg.in/check.v1"
)

func (s *ModelSuite) TestJavascriptValidation(c *gc.C) {
	for i, t := range []struct {
		should       string
		data         string
		givenVM      *otto.Otto
		expectErrors aperrors.Errors
	}{{
		should:       "not return an error for a valid transformation",
		data:         "var foo = \"test\"",
		expectErrors: aperrors.Errors{},
	}, {
		should: "return an error for an invalid transformation",
		data:   "}",
		expectErrors: aperrors.Errors{
			"after": []string{
				"(anonymous): Line 1:1 Unexpected token }",
			},
		},
	}, {
		should:       "not return an error for a valid transformation and a supplied vm",
		data:         "var foo = \"test\"",
		expectErrors: aperrors.Errors{},
		givenVM:      otto.New(),
	}, {
		should: "return an error for an invalid transformation and a supplied vm",
		data:   "}",
		expectErrors: aperrors.Errors{
			"after": []string{
				"(anonymous): Line 1:1 Unexpected token }",
			},
		},
		givenVM: otto.New(),
	}} {
		c.Logf("test %d: should %s", i, t.should)

		given := &model.ProxyEndpointTransformation{
			Type: model.ProxyEndpointTransformationTypeJS,
			Data: []byte(strconv.Quote(t.data)),
		}

		errors := given.Validate(t.givenVM)
		c.Check(errors, jc.DeepEquals, t.expectErrors)
	}
}
