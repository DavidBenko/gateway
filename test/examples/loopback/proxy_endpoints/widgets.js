function RestController() {}

RestController.prototype.handle = function(request) {
	if (request.method == "GET") {
		return this.get(request);
	} else if (request.method == "POST") {
		return this.post(request);
	}
}

RestController.prototype.get = function(request) {};
RestController.prototype.post = function(request) {};

var Endpoints = Endpoints || {};
Endpoints.Widgets = function(){};
Endpoints.Widgets.prototype = new RestController();
Endpoints.Widgets.prototype.constructor = Endpoints.Widgets;

Endpoints.Widgets.prototype.get = function(request) {
	var response = new AP.HTTP.Response();
	response.body = "Gotten!\n"
	return response;
}

Endpoints.Widgets.prototype.post = function(request) {
	var response = new AP.HTTP.Response();
	response.body = "Posted!\n"
	return response;
}
