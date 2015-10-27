package model_test

import (
	aperrors "gateway/errors"
	"gateway/model"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func testingComponents() map[string]*model.SharedComponent {
	return map[string]*model.SharedComponent{
		"bad-name": &model.SharedComponent{
			ProxyEndpointComponent: model.ProxyEndpointComponent{
				Type: model.ProxyEndpointComponentTypeJS,
			},
		},
		"bad-id": &model.SharedComponent{
			Name: "foo",
			ProxyEndpointComponent: model.ProxyEndpointComponent{
				Type:              model.ProxyEndpointComponentTypeJS,
				SharedComponentID: 5,
			},
		},
		"bad-name-and-id": &model.SharedComponent{
			ProxyEndpointComponent: model.ProxyEndpointComponent{
				Type:              model.ProxyEndpointComponentTypeJS,
				SharedComponentID: 5,
			},
		},
		"good-simple": &model.SharedComponent{
			Name: "foo",
			ProxyEndpointComponent: model.ProxyEndpointComponent{
				Type: model.ProxyEndpointComponentTypeJS,
			},
		},
	}
}

func (s *ModelSuite) TestSharedComponentValidate(c *gc.C) {
	for i, t := range []struct {
		should       string
		given        *model.SharedComponent
		expectErrors aperrors.Errors
	}{{
		should: "validate on name",
		given:  testingComponents()["bad-name"],
		expectErrors: aperrors.Errors{
			"name": []string{"must not be blank"},
		},
	}, {
		should: "validate on shared id",
		given:  testingComponents()["bad-id"],
		expectErrors: aperrors.Errors{
			"shared_component_id": []string{"must not be defined"},
		},
	}, {
		should: "validate on both",
		given:  testingComponents()["bad-name-and-id"],
		expectErrors: aperrors.Errors{
			"shared_component_id": []string{"must not be defined"},
			"name":                []string{"must not be blank"},
		},
	}, {
		should:       "validate an acceptable component",
		given:        testingComponents()["good-simple"],
		expectErrors: aperrors.Errors{},
	}} {
		c.Logf("test %d: should %s", i, t.should)
		c.Assert(t.given, gc.NotNil)
		c.Check(t.given.Validate(), jc.DeepEquals, t.expectErrors)
	}
}
