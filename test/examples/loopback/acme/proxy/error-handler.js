/**
 * Things can go poorly. This is an example that shows that Gateway
 * reports back errors to the JS so that developers can handle them
 * in a business-appropriate way.
 *
 * $ curl localhost:5000/error
 * Error! Error!
 *
 */
Acme.Proxy.ErrorHandler = function() {
	this.handle = function(proxyRequest) {
		var request = new AP.HTTP.Request();
		request.url = "i am not a url"

		AP.log("Calling out to external service");
		var proxyResponse = AP.makeRequest(request);

		if (proxyResponse.error) {
			AP.log("Error on proxied request: " + proxyResponse.error);
			var response = new AP.HTTP.Response();
			response.statusCode = 500;
			response.body = "Error! Error!\n";
			return response;
		}

		return proxyResponse;
	};
}
