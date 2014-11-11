var AP = AP || {};

/**
 * Environment contains accessor methods for retrieving values from the
 * environment the proxy is currently running in.
 *
 * @namespace
 */
AP.Environment = AP.Environment || {};

/**
 * Get a value from the environment.
 *
 * @param {string} key The key of the value to fetch
 */
AP.Environment.get = function(key) {
  return __ap_environment_get(key);
}
