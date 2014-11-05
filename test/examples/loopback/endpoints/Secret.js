/**
 * Returns success only if the right secret was passed in the body.
 *
 * curl -d not_the_password localhost:5000/secret
 * Denied
 * 
 * $ curl -d password localhost:5000/secret
 * Accepted
 * 
 */
function main(request) {
	var response = new AP.HTTP.Response();
	if (request.body != 'password') {
		response.statusCode = 400;
		response.body = "Denied";
	} else {
		response.body = "Accepted";
	}
	return response;
}
