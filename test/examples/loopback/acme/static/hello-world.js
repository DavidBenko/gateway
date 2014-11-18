/**
 * Basic functionality! It's alive!
 *
 * $ curl localhost:5000
 * Hello, world!
 *
 */
Acme.Static.HelloWorld = function() {
	this.handle = function(request) {
		var response = new AP.HTTP.Response();
		response.body = "Hello, world!\n";
		return response;
	};
};
