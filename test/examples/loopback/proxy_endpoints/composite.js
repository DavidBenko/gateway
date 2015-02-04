/**
 * This endpoint executes two requests simultaneously, and when they
 * both return it joins their data and sends back the result.
 *
 * This is the 1-N mapping case generally, and I think it's a pretty
 * powerful use case.
 *
 * $ curl localhost:5000/composite
 * {
 *    "bar": "baz",
 *    "foo": "baf"
 * }
 */
Acme.Proxy.Composite = function() {
	this.handle = function(proxyRequest) {
		// Build two requests
		var first = new AP.HTTP.Request();
		first.url = "http://localhost:5000/foo";

		var second = new AP.HTTP.Request();
		second.url = "http://localhost:5000/bar";

		// Execute both simultaneously
		var responses = AP.makeRequests([first, second]);

		// Combine them
		var composite = {};
		for (index in responses) {
			var response = responses[index];
			var responseData = JSON.parse(response.body);
			for (key in responseData) {
				composite[key] = responseData[key];
			}
		}

		// Send back combined data
		var response = new AP.HTTP.Response();
		response.setJSONBodyPretty(composite);
		return response;
	};
}
