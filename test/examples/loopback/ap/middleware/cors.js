var AP = AP || {};
AP.Middleware = AP.Middleware || {};
AP.Middleware.CORS = function(next) {
	this.next = next;
}

AP.Middleware.CORS.prototype.handle = function(request) {
	var response = this.next.handle(request);
	response.headers = response.headers || {};
	response.headers["Access-Control-Allow-Origin"] = "*";
	return response;
}
