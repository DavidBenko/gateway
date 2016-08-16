package model

import (
	"fmt"
	aperrors "gateway/errors"
	apsql "gateway/sql"
)

const (
	apiExportCurrentVersion int64  = 1
	quote                   string = `"`
)

/***
 * Notes on import/export
 *
 * - Exported data is is "sanitized" mostly by removing IDs by manually setting
 *   them to 0 with the json "omitempty" directive attached to the struct.
 * - For indices in relationships, we use 1-indexing. This is so that "omitempty"
 *   still works intuitively with 0 values.
 */

// FindAPIForAccountIDForExport returns the full API definition ready for import.
func FindAPIForAccountIDForExport(db *apsql.DB, id, accountID int64) (*API, error) {
	api, err := FindAPIForAccountID(db, id, accountID)
	if err != nil {
		return nil, aperrors.NewWrapped("Finding API", err)
	}
	api.ID = 0
	api.ExportVersion = apiExportCurrentVersion

	api.Environments, err = AllEnvironmentsForAPIIDAndAccountID(db, id, accountID)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching environments", err)
	}
	environmentsIndexMap := make(map[int64]int)
	for index, environment := range api.Environments {
		environmentsIndexMap[environment.ID] = index + 1
		environment.APIID = 0
		environment.ID = 0
	}

	api.EndpointGroups, err = AllEndpointGroupsForAPIIDAndAccountID(db, id, accountID)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching endpoint groups", err)
	}
	endpointGroupsIndexMap := make(map[int64]int)
	for index, endpointGroup := range api.EndpointGroups {
		endpointGroupsIndexMap[endpointGroup.ID] = index + 1
		endpointGroup.APIID = 0
		endpointGroup.ID = 0
	}

	api.Libraries, err = AllLibrariesForAPIIDAndAccountID(db, id, accountID)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching libraries", err)
	}
	for _, library := range api.Libraries {
		library.APIID = 0
		library.ID = 0
	}

	api.RemoteEndpoints, err = AllRemoteEndpointsForAPIIDAndAccountID(db, id, accountID)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching remote endpoints", err)
	}
	remoteEndpointsIndexMap := make(map[int64]int)
	environmentDataIndexMap := make(map[int64]int)
	for index, endpoint := range api.RemoteEndpoints {
		remoteEndpointsIndexMap[endpoint.ID] = index + 1
		endpoint.APIID = 0
		endpoint.ID = 0
		for envIndex, envData := range endpoint.EnvironmentData {
			envData.ExportEnvironmentIndex = environmentsIndexMap[envData.EnvironmentID]
			envData.EnvironmentID = 0
			envData.Links = nil
			environmentDataIndexMap[envData.ID] = envIndex + 1
			envData.ID = 0
			envData.RemoteEndpointID = 0
		}
		if endpoint.Soap != nil && endpoint.Soap.Wsdl != "" {
			if err := endpoint.encodeWsdlForExport(); err != nil {
				return nil, aperrors.NewWrapped("Encoding wsdl for export", err)
			}
		}
	}

	stripTransformationIDs := func(transformations []*ProxyEndpointTransformation) {
		for _, t := range transformations {
			t.ID = 0
		}
	}

	sharedComponents, err := AllSharedComponentsForAPIIDAndAccountID(
		db, id, accountID,
	)
	if err != nil {
		return nil, aperrors.NewWrapped("fetching shared components", err)
	}
	sharedComponentIndexMap := make(map[int64]int)
	for index, sh := range sharedComponents {
		sh.APIID = 0
		sh.ProxyEndpointComponentReferenceID = nil
		sh.ProxyEndpointComponent.ID = 0
		// Store the 1-indexed number of the SharedComponent.
		sharedComponentIndexMap[sh.ID] = index + 1
		sh.ID = 0

		stripTransformationIDs(sh.BeforeTransformations)
		stripTransformationIDs(sh.AfterTransformations)

		for _, call := range sh.AllCalls() {
			call.ID = 0
			call.RemoteEndpoint = nil
			stripTransformationIDs(call.BeforeTransformations)
			stripTransformationIDs(call.AfterTransformations)
			call.ExportRemoteEndpointIndex =
				remoteEndpointsIndexMap[call.RemoteEndpointID]
			call.RemoteEndpointID = 0
		}
	}
	api.SharedComponents = sharedComponents

	// Very much room for optimization
	proxyEndpointsIndexMap := make(map[int64]int)
	proxyEndpoint := ProxyEndpoint{
		AccountID: accountID,
		APIID:     id,
	}
	api.ProxyEndpoints, err = proxyEndpoint.All(db)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching proxy endpoints", err)
	}
	for i, ep := range api.ProxyEndpoints {
		endpoint, err := ep.Find(db)
		if err != nil {
			return nil, aperrors.NewWrapped("Fetching proxy endpoint", err)
		}
		proxyEndpointsIndexMap[endpoint.ID] = i + 1
		endpoint.APIID = 0
		endpoint.ID = 0
		if endpoint.EndpointGroupID != nil {
			endpoint.ExportEndpointGroupIndex = endpointGroupsIndexMap[*endpoint.EndpointGroupID]
			endpoint.EndpointGroupID = nil
		}
		endpoint.ExportEnvironmentIndex = environmentsIndexMap[endpoint.EnvironmentID]
		endpoint.EnvironmentID = 0
		for _, c := range endpoint.Components {
			c.ID = 0

			if id := c.SharedComponentID; id != nil {
				index := sharedComponentIndexMap[*id]
				c.ExportSharedComponentIndex = &index
				c.SharedComponentID = nil
			}

			stripTransformationIDs(c.BeforeTransformations)
			stripTransformationIDs(c.AfterTransformations)
			for _, call := range c.AllCalls() {
				call.ID = 0
				call.RemoteEndpoint = nil
				stripTransformationIDs(call.BeforeTransformations)
				stripTransformationIDs(call.AfterTransformations)
				call.ExportRemoteEndpointIndex = remoteEndpointsIndexMap[call.RemoteEndpointID]
				call.RemoteEndpointID = 0
			}
			c.ProxyEndpointComponentReferenceID = nil
		}

		api.ProxyEndpoints[i] = endpoint
	}

	schema := ProxyEndpointSchema{AccountID: accountID, APIID: id}
	api.ProxyEndpointSchemas, err = schema.All(db)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching proxy endpoint schemas", err)
	}
	for _, schema := range api.ProxyEndpointSchemas {
		schema.ExportProxyEndpointIndex = proxyEndpointsIndexMap[schema.ProxyEndpointID]
		schema.APIID = 0
		schema.ProxyEndpointID = 0
		schema.ID = 0
	}

	pad := ScratchPad{AccountID: accountID, APIID: id}
	api.ScratchPads, err = pad.All(db)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching scratch pads", err)
	}
	for _, pad := range api.ScratchPads {
		pad.ID = 0
		pad.ExportEnvironmentDataIndex = environmentDataIndexMap[pad.EnvironmentDataID]
		pad.EnvironmentDataID = 0
	}

	return api, nil
}

// Import imports any supported version of an API definition
func (a *API) Import(tx *apsql.Tx) (err error) {
	defer func() {
		a.ExportVersion = 0
	}()

	switch a.ExportVersion {
	case 1:
		return a.ImportV1(tx)
	default:
		return fmt.Errorf("Export version %d is not supported", a.ExportVersion)
	}
}

// ImportV1 imports the whole API definition in v1 format
func (a *API) ImportV1(tx *apsql.Tx) (err error) {
	environmentsIDMap := make(map[int]int64)
	for index, environment := range a.Environments {
		environment.AccountID = a.AccountID
		environment.UserID = a.UserID
		environment.APIID = a.ID
		err = environment.Insert(tx)
		if err != nil {
			return aperrors.NewWrapped("Inserting environment", err)
		}
		environmentsIDMap[index+1] = environment.ID
	}

	endpointGroupsIDMap := make(map[int]int64)
	for index, endpointGroup := range a.EndpointGroups {
		endpointGroup.AccountID = a.AccountID
		endpointGroup.UserID = a.UserID
		endpointGroup.APIID = a.ID
		err = endpointGroup.Insert(tx)
		if err != nil {
			return aperrors.NewWrapped("Inserting endpoint group", err)
		}
		endpointGroupsIDMap[index+1] = endpointGroup.ID
	}

	for _, library := range a.Libraries {
		library.AccountID = a.AccountID
		library.UserID = a.UserID
		library.APIID = a.ID
		err = library.Insert(tx)
		if err != nil {
			return aperrors.NewWrapped("Inserting library", err)
		}
	}

	remoteEndpointsIDMap := make(map[int]int64)
	type environmentDataIDs struct {
		remoteEndpointID, remoteEndpointEnvironmentDataID int64
	}
	environmentDataIDMap, environmentDataIndex := make(map[int]environmentDataIDs), 1
	for index, endpoint := range a.RemoteEndpoints {
		for _, envData := range endpoint.EnvironmentData {
			envData.EnvironmentID = environmentsIDMap[envData.ExportEnvironmentIndex]
			envData.ExportEnvironmentIndex = 0
		}

		endpoint.AccountID = a.AccountID
		endpoint.UserID = a.UserID
		endpoint.APIID = a.ID
		if vErr := endpoint.Validate(true); !vErr.Empty() {
			return fmt.Errorf("Unable to validate remote endpoint: %v", vErr)
		}
		err = endpoint.Insert(tx)
		if err != nil {
			return aperrors.NewWrapped("Inserting remote endpoint", err)
		}
		remoteEndpointsIDMap[index+1] = endpoint.ID
		for _, envData := range endpoint.EnvironmentData {
			environmentDataIDMap[environmentDataIndex] = environmentDataIDs{endpoint.ID, envData.ID}
			environmentDataIndex++
		}
	}

	// Map 1-indexed new SharedComponent indices to new SharedComponent IDs.
	sharedComponentsIDMap := make(map[int]int64)
	for index, sh := range a.SharedComponents {
		sh.AccountID, sh.UserID, sh.APIID = a.AccountID, a.UserID, a.ID

		for _, call := range sh.AllCalls() {
			call.RemoteEndpointID = remoteEndpointsIDMap[call.ExportRemoteEndpointIndex]
			call.ExportRemoteEndpointIndex = 0
		}

		if errs := sh.Validate(true); !errs.Empty() {
			var err error
			if base, ok := errs["base"]; ok {
				err = fmt.Errorf(
					"base validation error(s): %v", base,
				)
			} else {
				err = fmt.Errorf(
					"validation error(s): %v", errs,
				)
			}
			return aperrors.NewWrapped(
				"Inserting shared component", err,
			)
		}

		// Map API export IDs to new IDs.
		if err := sh.Insert(tx); err != nil {
			return aperrors.NewWrapped("Inserting shared component", err)
		}

		sharedComponentsIDMap[index+1] = sh.ID
	}

	proxyEndpointsIDMap := make(map[int]int64)
	for index, endpoint := range a.ProxyEndpoints {
		for _, c := range endpoint.Components {
			for _, call := range c.AllCalls() {
				call.RemoteEndpointID = remoteEndpointsIDMap[call.ExportRemoteEndpointIndex]
				call.ExportRemoteEndpointIndex = 0
			}
			if id := c.ExportSharedComponentIndex; id != nil {
				// If there is a non-nil SharedComponent index
				// export, assign the newly inserted
				// SharedComponent's ID here and nil out the
				// exported index.
				newID := sharedComponentsIDMap[*id]
				c.SharedComponentID = &newID
				c.ExportSharedComponentIndex = nil
			}
		}

		endpoint.EnvironmentID = environmentsIDMap[endpoint.ExportEnvironmentIndex]
		endpoint.ExportEnvironmentIndex = 0

		if endpoint.ExportEndpointGroupIndex != 0 {
			id := endpointGroupsIDMap[endpoint.ExportEndpointGroupIndex]
			endpoint.EndpointGroupID = &id
			endpoint.ExportEndpointGroupIndex = 0
		}

		endpoint.AccountID = a.AccountID
		endpoint.UserID = a.UserID
		endpoint.APIID = a.ID
		if err := endpoint.Insert(tx); err != nil {
			return aperrors.NewWrapped("Inserting proxy endpoint", err)
		}
		proxyEndpointsIDMap[index+1] = endpoint.ID
	}

	for _, schema := range a.ProxyEndpointSchemas {
		schema.AccountID = a.AccountID
		schema.UserID = a.UserID
		schema.APIID = a.ID
		schema.ProxyEndpointID = proxyEndpointsIDMap[schema.ExportProxyEndpointIndex]
		schema.ExportProxyEndpointIndex = 0
		err = schema.Insert(tx)
		if err != nil {
			return aperrors.NewWrapped("Inserting proxy endpoint schema", err)
		}
	}

	for _, pad := range a.ScratchPads {
		pad.AccountID = a.AccountID
		pad.UserID = a.UserID
		pad.APIID = a.ID
		ids := environmentDataIDMap[pad.ExportEnvironmentDataIndex]
		pad.RemoteEndpointID = ids.remoteEndpointID
		pad.EnvironmentDataID = ids.remoteEndpointEnvironmentDataID
		if vErr := pad.Validate(true); !vErr.Empty() {
			return fmt.Errorf("Unable to validate scratch pad: %v", vErr)
		}
		err = pad.Insert(tx)
		if err != nil {
			return aperrors.NewWrapped("Inserting scratch pad", err)
		}
	}

	return nil
}
