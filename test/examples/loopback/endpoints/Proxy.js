/**
 * Naked proxy. This just passes through the request to the remote service.
 * 
 * This case needs simplified -- far too many lines of code just to pass
 * something through unmodified. The work there is in creating a better
 * JavaScript API, or in allowing the proxyRequest to be an AP.HTTP.Reuqest.
 * 
 * $ curl localhost:5000/proxy
 * Hello, world!
 * 
 */
function main(proxyRequest) {
	var request = new AP.HTTP.Request();
	request.method = proxyRequest.method;
	request.headers = proxyRequest.headers;
	request.body = proxyRequest.body;
	request.url = "http://localhost:5000"
	return AP.makeRequest(request);
}
