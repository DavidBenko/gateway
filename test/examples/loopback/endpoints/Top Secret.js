/**
 * Static endpoint that's designed to be hidden. I don't actually
 * this this is secure, it's just to support the Workflow example.
 *
 * $ curl localhost:5000/topsecret
 * Super Secret Information
 * 
 */
function main(request) {
	var response = new AP.HTTP.Response();
	response.body = "Super Secret Information\n";
	return response;
}
