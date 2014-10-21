var AP = AP || {};

AP.Router = function() {
  this.routes = [];
}

AP.Router.prototype.routeData = function() {
  return JSON.stringify(this.routes);
}

AP.Router.prototype.newRoute = function() {
  var route = new AP.Route();
  this.routes.push(route);
  return route;
}

AP.Router.prototype.httpMethod = function(method, path, name) {
  return this.newRoute().httpMethod(method, path, name);
}

AP.Router.prototype.get = function(path, name) {
  return this.newRoute().get(path, name);
}

AP.Router.prototype.post = function(path, name) {
  return this.newRoute().post(path, name);
}

AP.Router.prototype.put = function(path, name) {
  return this.newRoute().put(path, name);
}

AP.Router.prototype.patch = function(path, name) {
  return this.newRoute().patch(path, name);
}

AP.Router.prototype.delete = function(path, name) {
  return this.newRoute().delete(path, name);
}

AP.Router.prototype.method = function(method) {
  return this.newRoute().method(method);
}

AP.Router.prototype.methods = function() {
  var route = this.newRoute();
  return this.newRoute().methods.apply(route, arguments);
}

AP.Router.prototype.path = function(path) {
  return this.newRoute().path(path);
}

AP.Router.prototype.name = function(name) {
  return this.newRoute().name(name);
}

var router = new AP.Router();
