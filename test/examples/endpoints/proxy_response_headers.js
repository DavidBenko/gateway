function main(proxyRequest) {
	var response = new AP.HTTP.Response();
	response.statusCode = 200;
	response.body = "I can haz headres?\n";
	response.headers = {
		"X-Custom1": "Foo",
		"X-Custom2": ["Bar", "Baz"]
	}
	return response;
}
