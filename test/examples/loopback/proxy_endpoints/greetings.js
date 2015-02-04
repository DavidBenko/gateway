/**
 * This is a semi-static endpoint that sends back a random greeting.
 * Its real function is to show how libraries can be used. All the
 * 'work' for this endpoint is specified in the Greetings library,
 * and can be reused across endpoints.
 *
 * $ curl localhost:5000/greetings
 * {
 *    "greeting": "How's it going?"
 * }
 * $ curl localhost:5000/greetings
 * {
 *    "greeting": "Greetings"
 * }
 * $ curl localhost:5000/greetings
 * {
 *    "greeting": "Sup?"
 * }
 *
 */

include("lib/Greetings");

Acme.Proxy.Greetings = function() {
	this.handle = function(request) {
		var response = new Greetings.Response();
		response.setBody(RandomGreeting());
		return response;
	};
}
