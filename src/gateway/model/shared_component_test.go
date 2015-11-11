package model_test

import (
	aperrors "gateway/errors"
	"gateway/model"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func (s *ModelSuite) TestSharedComponentValidate(c *gc.C) {
	for i, t := range []struct {
		should        string
		givenName     string
		givenType     string
		givenRefID    *int64
		givenSharedID *int64
		isInsert      bool
		expectErrors  aperrors.Errors
	}{{
		should:    "validate on name",
		givenType: model.ProxyEndpointComponentTypeJS,
		expectErrors: aperrors.Errors{
			"name": []string{"must not be blank"},
		},
	}, {
		should:        "validate on shared id",
		givenName:     "foo",
		givenType:     model.ProxyEndpointComponentTypeJS,
		givenSharedID: new(int64),
		expectErrors: aperrors.Errors{
			"shared_component_id": []string{"must not be defined"},
		},
	}, {
		should:        "validate on reference id",
		givenName:     "foo",
		givenType:     model.ProxyEndpointComponentTypeJS,
		givenSharedID: new(int64),
		givenRefID:    new(int64),
		expectErrors: aperrors.Errors{
			"shared_component_id": []string{"must not be defined"},
			"proxy_endpoint_component_reference_id": []string{
				"must not be defined",
			},
		},
	}, {
		should:        "validate on name and shared id",
		givenType:     model.ProxyEndpointComponentTypeJS,
		givenSharedID: new(int64),
		expectErrors: aperrors.Errors{
			"shared_component_id": []string{"must not be defined"},
			"name":                []string{"must not be blank"},
		},
	}, {
		should:        "validate on name, type, and shared id",
		givenSharedID: new(int64),
		expectErrors: aperrors.Errors{
			"shared_component_id": []string{"must not be defined"},
			"type": []string{
				"must be one of 'single', or 'multi', or 'js'",
			},
			"name": []string{"must not be blank"},
		},
	}, {
		should:       "validate an acceptable component",
		givenName:    "foo",
		givenType:    model.ProxyEndpointComponentTypeJS,
		expectErrors: aperrors.Errors{},
	}} {
		// We'll be catching a panic
		c.Logf("test %d: should %s", i, t.should)
		given := &model.SharedComponent{
			Name: t.givenName,
			ProxyEndpointComponent: model.ProxyEndpointComponent{
				Type:                              t.givenType,
				SharedComponentID:                 t.givenSharedID,
				ProxyEndpointComponentReferenceID: t.givenRefID,
			},
		}
		// Doesn't matter whether isInsert.
		errors := given.Validate(t.isInsert)
		c.Check(errors, jc.DeepEquals, t.expectErrors)
	}
}
