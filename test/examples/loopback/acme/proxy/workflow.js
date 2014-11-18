/**
 * This endpoint executes one request to check the user's password,
 * and then it conditionally performs a secondary action.
 *
 * This highlights the "workflow" use case, which I think is a
 * powerful feature. It's also something that is straightforward
 * to communicate in code, but harder to communicate in a web
 * interface without a lot of complexity (and graphical process flows).
 *
 * $ curl localhost:5000/workflow?secret=something
 * 'something' is not the secret.
 *
 * $ curl localhost:5000/workflow?secret=password
 * Super Secret Information
 *
 */
Acme.Proxy.Workflow = function() {
	this.handle = function(proxyRequest) {
		var secret = proxyRequest.params["secret"];

		// Check secret
		AP.log("Checking password...");
		var request = new AP.HTTP.Request();
		request.method = "POST";
		request.url = "http://localhost:5000/secret";
		request.body = secret;
		var passwordResponse = AP.makeRequest(request);

		// Make a second query only if the first one determined success
		if (passwordResponse.statusCode == 200) {
			AP.log("User knew password, continuing to confirmation");
			var finalRequest = new AP.HTTP.Request();
			finalRequest.headers["X-Sharedsecret"] = "12345";
			finalRequest.url = "http://localhost:5000/topsecret";
			return AP.makeRequest(finalRequest);
		} else {
			AP.log("User did not know password, denying further results");
			var response = new AP.HTTP.Response();
			response.statusCode = 401;
			response.body = "'" + secret + "' is not the secret.\n";
			return response;
		}
	};
}
