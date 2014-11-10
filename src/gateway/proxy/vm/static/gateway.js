/**
 * AP is the root namespace for all Gateway-provided functionality.
 *
 * @namespace
 */
var AP = AP || {};

/**
 * Logs the line to the Gateway log.
 *
 * The log line is tied to the individual request via a unique identifier and
 * is further keyed as "[proxy]" and "[user]" to differentiate the log line
 * from Gateway-created messages.
 *
 * For example,
 *     AP.log("Hello, Gateway!");
 *
 * Would produce a log line that looks like:
 *     2014/11/07 16:18:11.190725 [proxy] [req bc2645aa-2d51-4540-a9a4-8fbe82bac692] [user] Hello, Gateway!
 *
 * @param {string} line
 */
AP.log = function(line) {
  __ap_log(line);
}

/**
 * Makes a single request from the proxy.
 *
 * @param request The request the proxy should make
 * @return The response from the request
 */
AP.makeRequest = function(request) {
  return this.makeRequests([request])[0];
}

/**
 * Makes multiple concurrent requests from the proxy.
 *
 * @param {Array} requests An Array of requests the proxy should make
 * @return {Array} An array of responses from the request, in request order
 */
AP.makeRequests = function(requests) {
  var typedRequests = [];
  var numRequests = requests.length;
  for (var i = 0; i < numRequests; i++) {
    var request = requests[i];
    typedRequests.push([request.__ap_type, JSON.stringify(request)]);
  }
  var rawResponse = __ap_makeRequests(JSON.stringify(typedRequests));
  return JSON.parse(rawResponse);
}
