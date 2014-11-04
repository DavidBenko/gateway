function main(proxyRequest) {
	var secret = proxyRequest.params["secret"];
	
	// Check secret
	var request = new AP.HTTP.Request();
	request.method = "POST";
	request.url = "http://localhost:4567/secret";
	request.body = secret;
	AP.log("Checking password...");
	var passwordResponse = AP.makeRequest(request);
		
	// Make a second query only if the first one determined success
	if (passwordResponse.body == "true") {
		AP.log("User knew password, continuing to confirmation");
		var finalRequest = new AP.HTTP.Request();
		finalRequest.method = "GET";
		finalRequest.url = "http://localhost:4567/topsecret";
		return AP.makeRequest(finalRequest);
	} else {
		AP.log("User did not know password, denying further results");
		var response = new AP.HTTP.Response();
		response.statusCode = 401;
		response.body = "'" + secret + "' is not the secret.\n";
		return response;
	}
}

