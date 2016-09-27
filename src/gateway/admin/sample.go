package admin

import (
	"gateway/model"
	apsql "gateway/sql"
	"gateway/stats"
	"gateway/stats/sql"
)

// BeforeValidate makes sure any requested variables and constraints are within
// the user's permission to request.  It also injects a constraint on user_id to
// ensure that only samples for the given user are selected.
func (c *SamplesController) BeforeValidate(
	sample *model.Sample, tx *apsql.Tx,
) error {
	if err := sample.ValidateConstraints(tx); err != nil {
		return err
	}

	return nil
}

// QueryStats returns the results of valid variables and constraints in the given Sample query
func (c *SamplesController) QueryStats(
	sample *model.Sample, tx *apsql.Tx,
) (stats.Result, error) {
	if err := c.BeforeValidate(sample, tx); err != nil {
		return nil, err
	}
	sq := &sql.SQL{DB: tx.DB.DB}
	results, e := sq.Sample(
		sample.Constraints,
		sample.Variables...,
	)
	if e != nil {
		return nil, e
	}
	return results, nil
}
