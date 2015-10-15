package model_test

import (
	"gateway/model"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func (s *ModelSuite) TestSharedComponentValidate(c *gc.C) {
	for i, t := range []struct {
		should       string
		given        *model.SharedComponent
		expectErrors model.Errors
	}{{
		should: "validate on name",
		given: &model.SharedComponent{
			ProxyEndpointComponent: model.ProxyEndpointComponent{
				Type: model.ProxyEndpointComponentTypeJS,
			},
		},
		expectErrors: model.Errors{"name": []string{"must not be blank"}},
	}, {
		should: "validate on shared id",
		given: &model.SharedComponent{
			Name: "foo",
			ProxyEndpointComponent: model.ProxyEndpointComponent{
				Type:              model.ProxyEndpointComponentTypeJS,
				SharedComponentID: 5,
			},
		},
		expectErrors: model.Errors{"shared_component_id": []string{"must not be defined"}},
	}, {
		should: "validate on both",
		given: &model.SharedComponent{
			ProxyEndpointComponent: model.ProxyEndpointComponent{
				Type:              model.ProxyEndpointComponentTypeJS,
				SharedComponentID: 5,
			},
		},
		expectErrors: model.Errors{
			"shared_component_id": []string{"must not be defined"},
			"name":                []string{"must not be blank"},
		},
	}, {
		should: "validate an acceptable component",
		given: &model.SharedComponent{
			Name: "foo",
			ProxyEndpointComponent: model.ProxyEndpointComponent{
				Type: model.ProxyEndpointComponentTypeJS,
			},
		},
		expectErrors: model.Errors{},
	}} {
		c.Logf("test %d: should %s", i, t.should)
		c.Assert(t.given, gc.NotNil)
		c.Check(t.given.Validate(), jc.DeepEquals, t.expectErrors)
	}
}
