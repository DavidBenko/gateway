/**
 * Static endpoint to feed into the Composite endpoint.
 *
 * $ curl localhost:5000/bar
 * {
 *    "bar": "baz"
 * }
 *
 */
Acme.Static.Bar = function() {
	this.handle = function(request) {
		var response = new AP.HTTP.Response();
		response.setJSONBodyPretty({bar: "baz"});
		return response;
	};
}
