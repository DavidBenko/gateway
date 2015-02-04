/**
 * Naked proxy. This just passes through the request to the remote service,
 * changing the URL dynamically.
 *
 * $ curl localhost:5000/proxy
 * Hello, world!
 * $ curl localhost:5000/proxyEcho -d "foobar"
 * foobar
 *
 */
Acme.Proxy.Proxy = function() {
	this.handle = function(proxyRequest) {
		var request = new AP.HTTP.Request(proxyRequest);
		request.url = "http://localhost:5000" + proxyRequest.path.replace("proxy", "");
		return AP.makeRequest(request);
	};
}
