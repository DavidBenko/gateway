/**
 * Basic functionality! It's alive!
 *
 * $ curl localhost:5000
 * Hello, world!
 * 
 */
function main(request) {
	var response = new AP.HTTP.Response();
	response.body = "Hello, world!\n";
	return response;
}
