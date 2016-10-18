package admin

import (
	"gateway/model"
	apsql "gateway/sql"
	"gateway/stats"
	statssql "gateway/stats/sql"
)

// BeforeValidate makes sure any requested variables and constraints are within
// the user's permission to request.  It also injects a constraint on user_id to
// ensure that only samples for the given user are selected.
func (c *SamplesController) BeforeValidate(
	sample *model.Sample, db *apsql.DB,
) error {
	if err := sample.ValidateConstraints(db); err != nil {
		return err
	}

	return nil
}

// QueryStats returns the results of valid variables and constraints in the given Sample query
func (c *SamplesController) QueryStats(
	sample *model.Sample, db *apsql.DB, statsDb *statssql.SQL,
) (stats.Result, error) {
	if err := c.BeforeValidate(sample, db); err != nil {
		return nil, err
	}

	results, e := statsDb.Sample(
		sample.Constraints,
		sample.Variables...,
	)
	if e != nil {
		return nil, e
	}
	return results, nil
}
