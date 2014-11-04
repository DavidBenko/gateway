function main(request) {
	var response = new Greetings.Response();
	response.setBody(RandomGreeting());
	return response;
}
