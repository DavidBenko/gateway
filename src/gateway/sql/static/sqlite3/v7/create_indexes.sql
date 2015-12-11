-- apis
CREATE INDEX idx_apis_account_id ON apis(account_id);

-- endpoint_groups
CREATE INDEX idx_endpoint_groups_api_id ON endpoint_groups(api_id);

-- environments
CREATE INDEX idx_environments_api_id ON environments(api_id);

-- hosts
CREATE INDEX idx_hosts_api_id ON hosts(api_id);

-- libraries
CREATE INDEX idx_libraries_api_id ON libraries(api_id);

-- proxy_endpoints
CREATE INDEX idx_proxy_endpoints_api_id ON proxy_endpoints(api_id);
CREATE INDEX idx_proxy_endpoints_endpoint_group_id ON proxy_endpoints(endpoint_group_id);
CREATE INDEX idx_proxy_endpoints_environment_id ON proxy_endpoints(environment_id);

-- proxy_endpoint_calls
CREATE INDEX idx_proxy_endpoint_calls_component_id ON proxy_endpoint_calls(component_id);
CREATE INDEX idx_proxy_endpoint_calls_remote_endpoint_id ON proxy_endpoint_calls(remote_endpoint_id);

-- proxy_endpoint_components
CREATE INDEX idx_proxy_endpoint_components_endpoint_id ON proxy_endpoint_components(endpoint_id);

-- proxy_endpoint_transformations
CREATE INDEX idx_proxy_endpoint_transformations_component_id ON proxy_endpoint_transformations(component_id);
CREATE INDEX idx_proxy_endpoint_transformations_call_id ON proxy_endpoint_transformations(call_id);

-- remote_endpoints
CREATE INDEX idx_remote_endpoints_api_id ON remote_endpoints(api_id);

-- proxy_endpoint_schemas
CREATE INDEX idx_proxy_endpoint_schemas_endpoint_id ON proxy_endpoint_schemas(endpoint_id);
CREATE INDEX idx_proxy_endpoint_schemas_request_schema_id ON proxy_endpoint_schemas(request_schema_id);
CREATE INDEX idx_proxy_endpoint_schemas_response_schema_id ON proxy_endpoint_schemas(response_schema_id);

-- proxy_endpoint_test_pairs
CREATE INDEX idx_proxy_endpoint_test_pairs_test_id ON proxy_endpoint_test_pairs(test_id);

-- proxy_endpoints_tests
CREATE INDEX idx_proxy_endpoint_tests_endpoint_id ON proxy_endpoint_tests(endpoint_id);

-- schemas
CREATE INDEX idx_schemas_api_id ON schemas(api_id);

-- soap_remote_endpoints
CREATE INDEX idx_soap_remote_endpoints_remote_endpoint_id ON soap_remote_endpoints(remote_endpoint_id);

-- users
CREATE INDEX idx_users_account_id ON users(account_id);
CREATE INDEX idx_users_token ON users(token);
CREATE INDEX idx_users_email ON users(email);

ANALYZE;
