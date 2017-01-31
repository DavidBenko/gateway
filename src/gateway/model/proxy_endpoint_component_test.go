package model_test

import (
	"fmt"
	aperrors "gateway/errors"
	"gateway/model"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func (s *ModelSuite) TestProxyEndpointComponentValidate(c *gc.C) {
	for i, t := range []struct {
		should        string
		givenType     string
		givenRefID    *int64
		givenSharedID *int64
		givenIsInsert bool
		expectErrors  aperrors.Errors
	}{{
		should:        "on insert, validate non-shared on type failure",
		givenIsInsert: true,
		expectErrors: aperrors.Errors{
			"type": []string{
				"must be one of 'single', or 'multi', or 'js'",
			},
		},
	}, {
		should:        "on insert, validate non-shared on type OK",
		givenIsInsert: true,
		givenType:     model.ProxyEndpointComponentTypeJS,
		expectErrors:  aperrors.Errors{},
	}, {
		should:        "on insert, validate non-shared on type OK",
		givenIsInsert: true,
		givenType:     model.ProxyEndpointComponentTypeSingle,
		expectErrors:  aperrors.Errors{},
	}, {
		should:        "on insert, validate non-shared on type OK",
		givenIsInsert: true,
		givenType:     model.ProxyEndpointComponentTypeMulti,
		expectErrors:  aperrors.Errors{},
	}, {
		should:        "on insert, shortcut validation if has shared",
		givenIsInsert: true,
		givenType:     "something-wrong",
		givenSharedID: new(int64),
		expectErrors:  aperrors.Errors{},
	}, {
		should:    "on update, validate non-shared on missing refID",
		givenType: model.ProxyEndpointComponentTypeJS,
		expectErrors: aperrors.Errors{
			"proxy_endpoint_component_reference_id": []string{
				"must not be undefined",
			},
		},
	}, {
		should: "on update, validate non-shared on type and missing refID",
		expectErrors: aperrors.Errors{
			"type": []string{
				"must be one of 'single', or 'multi', or 'js'",
			},
			"proxy_endpoint_component_reference_id": []string{
				"must not be undefined",
			},
		},
	}, {
		should:       "on update, validate non-shared on type OK",
		givenType:    model.ProxyEndpointComponentTypeJS,
		givenRefID:   new(int64),
		expectErrors: aperrors.Errors{},
	}, {
		should:       "on update, validate non-shared on type OK",
		givenType:    model.ProxyEndpointComponentTypeSingle,
		givenRefID:   new(int64),
		expectErrors: aperrors.Errors{},
	}, {
		should:       "on update, validate non-shared on type OK",
		givenType:    model.ProxyEndpointComponentTypeMulti,
		givenRefID:   new(int64),
		expectErrors: aperrors.Errors{},
	}, {
		should:        "on update, shortcut other validation if SharedComponentID",
		givenType:     "something-wrong",
		givenSharedID: new(int64),
		expectErrors: aperrors.Errors{
			"proxy_endpoint_component_reference_id": []string{
				"must not be undefined",
			},
		},
	}, {
		should:        "on update, shortcut other validation if SharedComponentID",
		givenType:     "something-wrong",
		givenSharedID: new(int64),
		givenRefID:    new(int64),
		expectErrors:  aperrors.Errors{},
	}} {
		c.Logf("test %d: should %s", i, t.should)

		given := &model.ProxyEndpointComponent{
			SharedComponentID:                 t.givenSharedID,
			ProxyEndpointComponentReferenceID: t.givenRefID,
			Type: t.givenType,
		}

		errors := given.Validate(t.givenIsInsert)
		c.Check(errors, jc.DeepEquals, t.expectErrors)
	}
}

func (s *ModelSuite) TestProxyEndpointComponentValidateJavascript(c *gc.C) {
	for i, t := range []struct {
		should       string
		given        string
		expectErrors aperrors.Errors
	}{{
		should:       "not return an error with empty javascript data",
		given:        "",
		expectErrors: aperrors.Errors{},
	}, {
		should: "return an error with invalid javascript data",
		given:  "}",
		expectErrors: aperrors.Errors{
			"data": []string{
				"(anonymous): Line 1:1 Unexpected token }",
			},
		},
	}} {
		c.Logf("test %d: should %s", i, t.should)

		component := &model.ProxyEndpointComponent{
			Data: []byte(fmt.Sprintf("\"%s\"", t.given)),
			ProxyEndpointComponentReferenceID: new(int64),
			Type: "js",
		}

		errors := component.Validate(true)
		c.Check(errors, jc.DeepEquals, t.expectErrors)
	}
}
