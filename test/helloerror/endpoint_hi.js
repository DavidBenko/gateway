function main(proxyRequest) {
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
}
