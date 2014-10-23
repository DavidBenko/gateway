function main(proxyRequest) {
	var request = new AP.HTTP.Request();
	request.url = "http://localhost:4567"
	request.headers = {
		"X-Custom1": "Foo",
		"X-Custom2": ["Bar", "Baz"]
	}
	
	var response = AP.makeRequest(request);
	AP.log(JSON.stringify(response));
	return response;
}
