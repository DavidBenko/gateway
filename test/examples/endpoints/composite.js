function main(proxyRequest) {
	// Build two requests
	var first = new AP.HTTP.Request();
	first.url = "http://localhost:4567/foo";
	
	var second = new AP.HTTP.Request();
	second.url = "http://localhost:4567/bar";
	
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
	response.statusCode = 200;
	response.body = JSON.stringify(composite);
	return response;
}

