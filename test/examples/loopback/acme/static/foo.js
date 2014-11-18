/**
 * Static endpoint to feed into the Composite endpoint.
 *
 * $ curl localhost:5000/foo
 * {
 *    "foo": "baf"
 * }
 *
 */
Acme.Static.Foo = function() {
	this.handle = function(request) {
		var response = new AP.HTTP.Response();
		response.setJSONBodyPretty({foo: "baf"});
		return response;
	};
}
