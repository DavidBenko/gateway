-- apis.account_id
CREATE INDEX idx_apis_account_id ON apis USING btree(account_id);

-- endpoint_groups.api_id
CREATE INDEX idx_endpoint_groups_api_id ON endpoint_groups USING btree(api_id);

-- environments.api_id
CREATE INDEX idx_environments_api_id ON environments USING btree(api_id);

-- hosts.api_id
CREATE INDEX idx_hosts_api_id ON hosts USING btree(api_id);

-- libraries.api_id
CREATE INDEX idx_libraries_api_id ON libraries USING btree(api_id);

-- proxy_endpoints.api_id
CREATE INDEX idx_proxy_endpoints_api_id ON proxy_endpoints USING btree(api_id);

-- proxy_endpoints.endpoint_group_id
CREATE INDEX idx_proxy_endpoints_endpoint_group_id ON proxy_endpoints USING btree(endpoint_group_id);

-- proxy_endpoints.environment_id
CREATE INDEX idx_proxy_endpoints_environment_id ON proxy_endpoints USING btree(environment_id);

-- proxy_endpoint_calls.component_id
CREATE INDEX idx_proxy_endpoint_calls_component_id ON proxy_endpoint_calls USING btree(component_id);

-- proxy_endpoint_calls.remote_endpoint_id
CREATE INDEX idx_proxy_endpoint_calls_remote_endpoint_id ON proxy_endpoint_calls USING btree(remote_endpoint_id);

-- proxy_endpoint_components.endpoint_id
CREATE INDEX idx_proxy_endpoint_components_endpoint_id ON proxy_endpoint_components USING btree(endpoint_id);

-- proxy_endpoint_transformations.component_id
CREATE INDEX idx_proxy_endpoint_transformations_component_id ON proxy_endpoint_transformations USING btree(component_id);

-- proxy_endpoint_transformations.call_id
CREATE INDEX idx_proxy_endpoint_transformations_call_id ON proxy_endpoint_transformations USING btree(call_id);

-- remote_endpoints.api_id
CREATE INDEX idx_remote_endpoints_api_id ON remote_endpoints USING btree(api_id);

-- proxy_endpoint_schemas.endpoint_id
CREATE INDEX idx_proxy_endpoint_schemas_endpoint_id ON proxy_endpoint_schemas USING btree(endpoint_id);

-- proxy_endpoint_schemas.request_schema_id
CREATE INDEX idx_proxy_endpoint_schemas_request_schema_id ON proxy_endpoint_schemas USING btree(request_schema_id);

-- proxy_endpoint_schemas.response_schema_id
CREATE INDEX idx_proxy_endpoint_schemas_response_schema_id ON proxy_endpoint_schemas USING btree(response_schema_id);

-- proxy_endpoint_test_pairs.test_id
CREATE INDEX idx_proxy_endpoint_test_pairs_test_id ON proxy_endpoint_test_pairs USING btree(test_id);

-- proxy_endpoints_tests.endpoint_id
CREATE INDEX idx_proxy_endpoint_tests_endpoint_id ON proxy_endpoint_tests USING btree(endpoint_id);

-- schemas.api_id
CREATE INDEX idx_schemas_api_id ON schemas USING btree(api_id);

-- soap_remote_endpoints.remote_endpoint_id
CREATE INDEX idx_soap_remote_endpoints_remote_endpoint_id ON soap_remote_endpoints USING btree(remote_endpoint_id);

-- users.account_id
CREATE INDEX idx_users_account_id ON users USING btree(account_id);

-- users.token
CREATE INDEX idx_users_token ON users USING btree(token);

-- users.email
CREATE INDEX idx_users_email ON users USING btree(email);

ANALYZE;



FOREIGN KEY("component_id") REFERENCES "proxy_endpoint_components"("id") ON DELETE CASCADE,
FOREIGN KEY("remote_endpoint_id") REFERENCES "remote_endpoints"("id") DEFERRABLE INITIALLY DEFERRED
