function main(request) {
	var response = new AP.HTTP.Response();
	response.statusCode = 205;
	response.body = JSON.stringify(request);
	return response;
}
