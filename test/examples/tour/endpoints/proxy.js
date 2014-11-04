function main(proxyRequest) {
	var request = new AP.HTTP.Request();
	request.url = "http://localhost:4567"
	
	AP.log("Calling out to external service");
	var proxyResponse = AP.makeRequest(request);

	AP.log("Returning: " + JSON.stringify(proxyResponse));
	var response = new AP.HTTP.Response();
	response.statusCode = 200;
	response.body = JSON.stringify(proxyResponse);
	return response;
}
