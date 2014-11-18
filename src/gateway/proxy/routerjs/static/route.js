var AP = AP || {};

AP.Route = function() {
  this._methods = [];
  this._schemes = [];
  this._host = null;
  this._path = null;
  this._pathPrefix = null;
  this._headers = null;
  this._queries = null;
  this._name = null;
}

AP.Route.prototype.httpMethod = function(method, path, name) {
  this._methods.push(method);
  this._path = path;
  this._name = name;
  return this;
}

AP.Route.prototype.get = function(path, name) {
  return this.httpMethod("GET", path, name);
}

AP.Route.prototype.post = function(path, name) {
  return this.httpMethod("POST", path, name);
}

AP.Route.prototype.put = function(path, name) {
  return this.httpMethod("PUT", path, name);
}

AP.Route.prototype.patch = function(path, name) {
  return this.httpMethod("PATCH", path, name);
}

AP.Route.prototype.delete = function(path, name) {
  return this.httpMethod("DELETE", path, name);
}

AP.Route.prototype.method = function(method) {
  this._methods.push(method);
  return this;
}

AP.Route.prototype.methods = function() {
  this._methods = Array.prototype.slice.call(arguments);
  return this;
}

AP.Route.prototype.path = function(path) {
  this._path = path;
  return this;
}

AP.Route.prototype.pathPrefix = function(pathPrefix) {
  this._pathPrefix = pathPrefix;
  return this;
}

AP.Route.prototype.host = function(host) {
  this._host = host;
  return this;
}

AP.Route.prototype.scheme = function(scheme) {
  this._schemes.push(scheme);
  return this;
}

AP.Route.prototype.schemes = function() {
  this.this = Array.prototype.slice.call(arguments);
  return this;
}

AP.Route.prototype.headers = function(headers) {
  this._headers = headers;
  return this;
}

AP.Route.prototype.queries = function(queries) {
  this._queries = queries;
  return this;
}

AP.Route.prototype.name = function(name) {
  this._name = name;
  return this;
}
