function main(request) {
	var response = new AP.HTTP.Response();
	response.setJSONBodyPretty(request);
	return response;
}
