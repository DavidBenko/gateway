/**
 * This example shows using environment values in proxy code.
 * 
 * $ curl localhost:5000/env
 * How am I? I'm fine.
 * And how are you?
 * 
 */
function main(request) {
    var response = new AP.HTTP.Response();
    var body = "How am I? I'm " + AP.Environment.get("mood") + ".\n";
	if (AP.Environment.get("shouldPrompt")) {
		body += "And how are you?\n";
	}
	response.body = body;
    return response;
}
