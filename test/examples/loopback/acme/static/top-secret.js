/**
 * Static endpoint that's designed to be hidden. In this case
 * it is designed not to be accessible without the presence of
 * a shared secret which only the proxy code should know.
 *
 * Actual real security mechanisms could be used instead;
 * this is just to support the Workflow example.
 *
 * curl localhost:5000/topsecret
 * {
 *   "error": "Access denied!"
 * }
 *
 * $ curl -H "X-Sharedsecret: 12345" localhost:5000/topsecret
 * Super Secret Information
 *
 */
Acme.Static.TopSecret = function() {
	this.handle = function(request) {
		if (request.headers["X-Sharedsecret"] != "12345") {
			var response = new AP.HTTP.Response();
			response.statusCode = 401;
			response.setJSONBodyPretty({error: "Access denied!"});
			return response;
		}

		var response = new AP.HTTP.Response();
		response.body = "Super Secret Information\n";
		return response;
	};
}
