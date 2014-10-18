function Router() {
  this.routes = [];
}

Router.prototype.routeData = function() {
  return JSON.stringify(this.routes);
}

Router.prototype.newRoute = function() {
  var route = new Route();
  this.routes.push(route);
  return route;
}

Router.prototype.httpMethod = function(method, path, name) {
  return this.newRoute().httpMethod(method, path, name);
}

Router.prototype.get = function(path, name) {
  return this.newRoute().get(path, name);
}

Router.prototype.post = function(path, name) {
  return this.newRoute().post(path, name);
}

Router.prototype.put = function(path, name) {
  return this.newRoute().put(path, name);
}

Router.prototype.patch = function(path, name) {
  return this.newRoute().patch(path, name);
}

Router.prototype.delete = function(path, name) {
  return this.newRoute().delete(path, name);
}

Router.prototype.method = function(method) {
  return this.newRoute().method(method);
}

Router.prototype.methods = function() {
  var route = this.newRoute();
  return this.newRoute().methods.apply(route, arguments);
}

Router.prototype.path = function(path) {
  return this.newRoute().path(path);
}

Router.prototype.name = function(name) {
  return this.newRoute().name(name);
}

var router = new Router();
