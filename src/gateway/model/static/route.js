function Route() {
  this._methods = [];
  this._path = null;
  this._name = null;
}

Route.prototype.httpMethod = function(method, path, name) {
  this._methods.push(method);
  this._path = path;
  this._name = name;
  return this;
}

Route.prototype.get = function(path, name) {
  return this.httpMethod("GET", path, name);
}

Route.prototype.post = function(path, name) {
  return this.httpMethod("POST", path, name);
}

Route.prototype.put = function(path, name) {
  return this.httpMethod("PUT", path, name);
}

Route.prototype.patch = function(path, name) {
  return this.httpMethod("PATCH", path, name);
}

Route.prototype.delete = function(path, name) {
  return this.httpMethod("DELETE", path, name);
}

Route.prototype.method = function(method) {
  this._methods.push(method);
  return this;
}

Route.prototype.methods = function() {
  this._methods = Array.prototype.slice.call(arguments);
  return this;
}

Route.prototype.path = function(path) {
  this._path = path;
  return this;
}

Route.prototype.name = function(name) {
  this._name = name;
  return this;
}
