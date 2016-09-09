package admin

import (
	"gateway/model"
	apsql "gateway/sql"
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
