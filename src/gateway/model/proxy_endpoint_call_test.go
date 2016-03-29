package model_test

import (
	aperrors "gateway/errors"
	"gateway/model"

	jc "github.com/juju/testing/checkers"
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

		errors := given.Validate()
		c.Check(errors, jc.DeepEquals, t.expectErrors)
	}
}
