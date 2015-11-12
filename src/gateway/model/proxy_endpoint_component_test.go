package model_test

import (
	aperrors "gateway/errors"
	"gateway/model"

	"github.com/jmoiron/sqlx/types"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func (s *ModelSuite) TestValidateNonShared(c *gc.C) {
	for i, t := range []struct {
		should       string
		givenType    string
		expectErrors aperrors.Errors
	}{{
		should: "validate non-shared on type failure",
		expectErrors: aperrors.Errors{
			"type": []string{
				"must be one of 'single', or 'multi', or 'js'",
			},
		},
	}, {
		should:       "validate non-shared on type OK",
		givenType:    model.ProxyEndpointComponentTypeJS,
		expectErrors: aperrors.Errors{},
	}, {
		should:       "validate non-shared on type OK",
		givenType:    model.ProxyEndpointComponentTypeSingle,
		expectErrors: aperrors.Errors{},
	}, {
		should:       "validate non-shared on type OK",
		givenType:    model.ProxyEndpointComponentTypeMulti,
		expectErrors: aperrors.Errors{},
	}} {
		c.Logf("test %d: should %s", i, t.should)
		given := &model.ProxyEndpointComponent{
			Type: t.givenType,
		}
		c.Check(given.Validate(), jc.DeepEquals, t.expectErrors)
	}
}

func (s *ModelSuite) TestValidateShared(c *gc.C) {
	sg, m, j := model.ProxyEndpointComponentTypeSingle,
		model.ProxyEndpointComponentTypeMulti,
		model.ProxyEndpointComponentTypeJS
	empty, nonEmpty := types.JsonText(`""`), types.JsonText(`something`)

	for i, t := range []struct {
		should           string
		givenType        string
		givenCall        *model.ProxyEndpointCall
		givenCalls       []*model.ProxyEndpointCall
		givenData        types.JsonText
		givenSharedType  string
		givenSharedCall  *model.ProxyEndpointCall
		givenSharedCalls []*model.ProxyEndpointCall
		givenSharedData  types.JsonText
		expectErrors     aperrors.Errors
	}{{
		should:          "validate on type failure",
		givenType:       j,
		givenSharedType: sg,
		givenData:       empty,
		expectErrors: aperrors.Errors{
			"type": []string{
				"must equal shared component's type",
			},
		},
	}, {
		should:          "validate single on Calls failure",
		givenType:       sg,
		givenSharedType: sg,
		givenData:       empty,
		givenCalls:      make([]*model.ProxyEndpointCall, 1),
		expectErrors: aperrors.Errors{
			"calls": []string{
				"type " + sg + " must not have multi calls",
			},
		},
	}, {
		should:          "validate single on Data failure",
		givenType:       sg,
		givenSharedType: sg,
		givenData:       nonEmpty,
		expectErrors: aperrors.Errors{
			"data": []string{
				"type " + sg + " must have empty js",
			},
		},
	}, {
		should:          "validate single on Data and Calls failures",
		givenType:       sg,
		givenSharedType: sg,
		givenCalls:      make([]*model.ProxyEndpointCall, 1),
		givenData:       nonEmpty,
		expectErrors: aperrors.Errors{
			"data": []string{
				"type " + sg + " must have empty js",
			},
			"calls": []string{
				"type " + sg + " must not have multi calls",
			},
		},
	}, {
		should:          "validate multi on Call failure",
		givenType:       m,
		givenSharedType: m,
		givenData:       empty,
		givenCall:       new(model.ProxyEndpointCall),
		expectErrors: aperrors.Errors{
			"call": []string{
				"type " + m + " must not have single call",
			},
		},
	}, {
		should:          "validate multi on Data failure",
		givenType:       m,
		givenSharedType: m,
		givenData:       nonEmpty,
		expectErrors: aperrors.Errors{
			"data": []string{
				"type " + m + " must have empty js",
			},
		},
	}, {
		should:          "validate multi on Data and Call failures",
		givenType:       m,
		givenSharedType: m,
		givenCall:       new(model.ProxyEndpointCall),
		givenData:       nonEmpty,
		expectErrors: aperrors.Errors{
			"data": []string{
				"type " + m + " must have empty js",
			},
			"call": []string{
				"type " + m + " must not have single call",
			},
		},
	}, {
		should:          "validate js on Call failure",
		givenType:       j,
		givenSharedType: j,
		givenData:       nonEmpty,
		givenCall:       new(model.ProxyEndpointCall),
		expectErrors: aperrors.Errors{
			"call": []string{
				"type " + j + " must not have single call",
			},
		},
	}, {
		should:          "validate js on Calls failure",
		givenType:       j,
		givenSharedType: j,
		givenData:       nonEmpty,
		givenCalls:      make([]*model.ProxyEndpointCall, 1),
		expectErrors: aperrors.Errors{
			"calls": []string{
				"type " + j + " must not have multi calls",
			},
		},
	}, {
		should:          "validate js on Call and Calls failures",
		givenType:       j,
		givenSharedType: j,
		givenData:       nonEmpty,
		givenCalls:      make([]*model.ProxyEndpointCall, 1),
		givenCall:       new(model.ProxyEndpointCall),
		expectErrors: aperrors.Errors{
			"call": []string{
				"type " + j + " must not have single call",
			},
			"calls": []string{
				"type " + j + " must not have multi calls",
			},
		},
	}, {
		should:          "validate shared on type OK",
		givenType:       j,
		givenSharedType: j,
		givenData:       empty,
		expectErrors:    aperrors.Errors{},
	}} {
		c.Logf("test %d: should %s", i, t.should)
		shared := &model.SharedComponent{
			ProxyEndpointComponent: model.ProxyEndpointComponent{
				Type:  t.givenSharedType,
				Call:  t.givenSharedCall,
				Calls: t.givenSharedCalls,
				Data:  t.givenSharedData,
			},
		}
		given := &model.ProxyEndpointComponent{
			Type:                  t.givenType,
			Call:                  t.givenCall,
			Calls:                 t.givenCalls,
			Data:                  t.givenData,
			SharedComponentID:     new(int64),
			SharedComponentHandle: shared,
		}
		c.Check(given.Validate(), jc.DeepEquals, t.expectErrors)
	}
}
