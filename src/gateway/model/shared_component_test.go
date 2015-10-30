package model_test

import (
	aperrors "gateway/errors"
	"gateway/model"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func (s *ModelSuite) TestSharedComponentValidate(c *gc.C) {
	five := new(int64)
	*five = 5

	for i, t := range []struct {
		should        string
		givenName     string
		givenType     string
		givenSharedID *int64
		expectErrors  aperrors.Errors
	}{{
		should:        "validate on name",
		givenType:     model.ProxyEndpointComponentTypeJS,
		givenSharedID: new(int64),
		expectErrors: aperrors.Errors{
			"name": []string{"must not be blank"},
		},
	}, {
		should:        "validate on shared id",
		givenName:     "foo",
		givenType:     model.ProxyEndpointComponentTypeJS,
		givenSharedID: five,
		expectErrors: aperrors.Errors{
			"shared_component_id": []string{"must not be defined"},
		},
	}, {
		should:        "validate on name and shared id",
		givenType:     model.ProxyEndpointComponentTypeJS,
		givenSharedID: five,
		expectErrors: aperrors.Errors{
			"shared_component_id": []string{"must not be defined"},
			"name":                []string{"must not be blank"},
		},
	}, {
		should:        "validate on name, type, and shared id",
		givenSharedID: five,
		expectErrors: aperrors.Errors{
			"type": []string{
				"must be one of 'single', or 'multi', or 'js'",
			},
			"shared_component_id": []string{"must not be defined"},
			"name":                []string{"must not be blank"},
		},
	}, {
		should:        "validate an acceptable component",
		givenName:     "foo",
		givenType:     model.ProxyEndpointComponentTypeJS,
		givenSharedID: new(int64),
		expectErrors:  aperrors.Errors{},
	}} {
		func() {
			// We'll be catching a panic
			c.Logf("test %d: should %s", i, t.should)
			given := &model.SharedComponent{
				Name: t.givenName,
				ProxyEndpointComponent: model.ProxyEndpointComponent{
					Type:              t.givenType,
					SharedComponentID: t.givenSharedID,
				},
			}
			if given.ProxyEndpointComponent.SharedComponentID != nil &&
				given.ProxyEndpointComponent.SharedComponentHandle == nil {
				// We expect Validate to panic if the Handle is
				// nil; handle the panic.
				defer handlePanic(c)
			}
			c.Check(given.Validate(), jc.DeepEquals, t.expectErrors)
		}()
	}
}

func handlePanic(c *gc.C) {
	e := recover()
	switch err := e.(type) {
	case error:
		c.Assert(err, gc.NotNil)
		c.Check(err, gc.ErrorMatches, `.* nil pointer dereference`)
	default:
		c.Logf("unexpected panic: %#v", e)
		c.FailNow()
	}
}
