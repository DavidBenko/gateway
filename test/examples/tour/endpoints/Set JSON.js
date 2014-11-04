function main(proxyRequest) {
	var request = new AP.HTTP.Request();
	request.method = "POST";
	request.url = "http://localhost:4567/echo"
	request.setJSONBody({"Foo": "Bar"});
	
	return AP.makeRequest(request);
}
