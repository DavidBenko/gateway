/**
 * Static endpoint to feed into the Composite endpoint.
 *
 * $ curl localhost:5000/bar
 * {
 *    "bar": "baz"
 * }
 * 
 */
function main(request) {
	var response = new AP.HTTP.Response();
	response.setJSONBodyPretty({bar: "baz"});
	return response;
}
