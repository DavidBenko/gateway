Acme.Middleware.Logger = function(next) {
	this.next = next;
}

Acme.Middleware.Logger.prototype.handle = function(request) {
	var start = new Date().getTime();
	var response = this.next.handle(request);
	var time = (new Date().getTime()) - start;
	AP.log("[middlware] Request took " + time + " ms in proxy code.");
	return response;
}
