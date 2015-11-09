package admin

import (
	"gateway/model"
	apsql "gateway/sql"
)

// BeforeValidate populates the SharedComponent handles of any of the
// ProxyEndpoint's Components which were inherited from SharedComponents.
func (c *ProxyEndpointController) BeforeValidate(
	p *model.ProxyEndpoint,
	tx *apsql.Tx,
) error {
	return p.PopulateSharedComponents(tx.DB)
}
