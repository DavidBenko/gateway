package model_test

import (
	aperrors "gateway/errors"
	"gateway/model"
	"strconv"

	jc "github.com/juju/testing/checkers"
	"github.com/robertkrimen/otto"
	gc "gopkg.in/check.v1"
)

func (s *ModelSuite) TestProxyEndpointCallValidate(c *gc.C) {
	for i, t := range []struct {
		should        string
		givenRemoteID int64
		expectErrors  aperrors.Errors
	}{{
		should: "validate missing RemoteEndpointID",
		expectErrors: aperrors.Errors{
			"remote_endpoint_id": []string{"must specify a remote endpoint"},
		},
	}, {
		should:        "validate acceptable Call",
		givenRemoteID: 1,
		expectErrors:  aperrors.Errors{},
	}} {
		c.Logf("test %d: should %s", i, t.should)

		given := &model.ProxyEndpointCall{
			RemoteEndpointID: t.givenRemoteID,
		}

		errors := given.Validate(nil)
		c.Check(errors, jc.DeepEquals, t.expectErrors)
	}
}

func (s *ModelSuite) TestProxyEndpontCallValidateJavascriptTransformations(c *gc.C) {
	for i, t := range []struct {
		should       string
		before       string
		after        string
		givenVM      *otto.Otto
		expectErrors aperrors.Errors
	}{{
		should:       "not return an error for valid before transformation",
		before:       "var foo = \"test\"",
		expectErrors: aperrors.Errors{},
	}, {
		should:       "not return an error for a valid after transformation",
		after:        "var foo = \"test\"",
		expectErrors: aperrors.Errors{},
	}, {
		should: "return an error for an invalid before transformation",
		before: "}",
		expectErrors: aperrors.Errors{
			"before": []string{
				"(anonymous): Line 1:1 Unexpected token }",
			},
		},
	}, {
		should: "return an error for an invalid after transformation",
		after:  "}",
		expectErrors: aperrors.Errors{
			"after": []string{
				"(anonymous): Line 1:1 Unexpected token }",
			},
		},
	}, {
		should:       "not return an error if nil is passed in place of a VM",
		before:       "var foo = \"test\"",
		givenVM:      nil,
		expectErrors: aperrors.Errors{},
	}, {
		should:       "not return an error if a VM is passed",
		before:       "var foo = \"test\"",
		givenVM:      otto.New(),
		expectErrors: aperrors.Errors{},
	}} {
		c.Logf("test %d: should %s", i, t.should)

		call := &model.ProxyEndpointCall{}
		call.RemoteEndpointID = 1

		if t.before != "" {
			transformation := &model.ProxyEndpointTransformation{}
			transformation.Type = "js"
			transformation.Before = true
			transformation.Data = []byte(strconv.Quote(t.before))
			call.BeforeTransformations = []*model.ProxyEndpointTransformation{transformation}
		}

		if t.after != "" {
			transformation := &model.ProxyEndpointTransformation{}
			transformation.Type = "js"
			transformation.Before = false
			transformation.Data = []byte(strconv.Quote(t.after))
			call.AfterTransformations = []*model.ProxyEndpointTransformation{transformation}
		}

		errors := call.Validate(t.givenVM)
		c.Check(errors, jc.DeepEquals, t.expectErrors)
	}
}
