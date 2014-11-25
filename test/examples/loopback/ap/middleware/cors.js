var AP = AP || {};
AP.Middleware = AP.Middleware || {};
AP.Middleware.CORS = function(next) {
	this.next = next;
}

AP.Middleware.CORS.prototype.handle = function(request) {
	var response = this.next.handle(request);
	var accessHeader = response.headers["Access-Control-Allow-Origin"]
	response.headers["Access-Control-Allow-Origin"] = accessHeader || "*";
	return response;
}
