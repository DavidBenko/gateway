package admin

import (
	"gateway/model"
	apsql "gateway/sql"
)

// BeforeInsert populates the SharedComponent handles of any of the
// ProxyEndpoint's Components which were inherited from SharedComponents.
func (c *ProxyEndpointController) BeforeInsert(
	p *model.ProxyEndpoint,
	tx *apsql.Tx,
) error {
	return p.PopulateSharedComponents(tx)
}

// BeforeUpdate populates the SharedComponent handles of any of the
// ProxyEndpoint's Components which were inherited from SharedComponents.
func (c *ProxyEndpointController) BeforeUpdate(
	p *model.ProxyEndpoint,
	tx *apsql.Tx,
) error {
	return p.PopulateSharedComponents(tx)
}
