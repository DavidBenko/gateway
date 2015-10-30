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
		givenType:       sg,
		givenSharedType: sg,
		expectErrors: aperrors.Errors{
			"type": []string{
				"must not override SharedComponent's type",
			},
		},
	}, {
		should:          "validate single on Calls failure",
		givenSharedType: model.ProxyEndpointComponentTypeSingle,
		givenCalls:      make([]*model.ProxyEndpointCall, 0),
		expectErrors: aperrors.Errors{
			"calls": []string{
				"type " + sg + " must not have multi calls",
			},
		},
	}, {
		should:          "validate single on Data failure",
		givenSharedType: model.ProxyEndpointComponentTypeSingle,
		givenData:       make(types.JsonText, 0),
		expectErrors: aperrors.Errors{
			"data": []string{
				"type " + sg + " must not have js",
			},
		},
	}, {
		should:          "validate single on Data and Calls failures",
		givenSharedType: model.ProxyEndpointComponentTypeSingle,
		givenCalls:      make([]*model.ProxyEndpointCall, 0),
		givenData:       make(types.JsonText, 0),
		expectErrors: aperrors.Errors{
			"data": []string{
				"type " + sg + " must not have js",
			},
			"calls": []string{
				"type " + sg + " must not have multi calls",
			},
		},
	}, {
		should:          "validate multi on Call failure",
		givenSharedType: model.ProxyEndpointComponentTypeMulti,
		givenCall:       new(model.ProxyEndpointCall),
		expectErrors: aperrors.Errors{
			"call": []string{
				"type " + m + " must not have single call",
			},
		},
	}, {
		should:          "validate multi on Data failure",
		givenSharedType: model.ProxyEndpointComponentTypeMulti,
		givenData:       make(types.JsonText, 0),
		expectErrors: aperrors.Errors{
			"data": []string{
				"type " + m + " must not have js",
			},
		},
	}, {
		should:          "validate multi on Data and Call failures",
		givenSharedType: model.ProxyEndpointComponentTypeMulti,
		givenCall:       new(model.ProxyEndpointCall),
		givenData:       make(types.JsonText, 0),
		expectErrors: aperrors.Errors{
			"data": []string{
				"type " + m + " must not have js",
			},
			"call": []string{
				"type " + m + " must not have single call",
			},
		},
	}, {
		should:          "validate JS on Call failure",
		givenSharedType: model.ProxyEndpointComponentTypeJS,
		givenCall:       new(model.ProxyEndpointCall),
		expectErrors: aperrors.Errors{
			"call": []string{
				"type " + j + " must not have single call",
			},
		},
	}, {
		should:          "validate js on Calls failure",
		givenSharedType: model.ProxyEndpointComponentTypeJS,
		givenCalls:      make([]*model.ProxyEndpointCall, 0),
		expectErrors: aperrors.Errors{
			"calls": []string{
				"type " + j + " must not have multi calls",
			},
		},
	}, {
		should:          "validate js on Call and Calls failures",
		givenSharedType: model.ProxyEndpointComponentTypeJS,
		givenCalls:      make([]*model.ProxyEndpointCall, 0),
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
		givenSharedType: model.ProxyEndpointComponentTypeSingle,
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
