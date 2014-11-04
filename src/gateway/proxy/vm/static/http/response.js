var AP = AP || {};
AP.HTTP = AP.HTTP || {};

AP.HTTP.Response = function() {
  this.statusCode = 200;
  this.body = null;
}

AP.HTTP.Response.prototype.setJSONBody = function(object) {
  this.headers = this.headers || {};
  this.headers["Content-Type"] = "application/json";
  this.body = JSON.stringify(object);
}

AP.HTTP.Response.prototype.setJSONBodyPretty = function(object) {
  this.headers = this.headers || {};
  this.headers["Content-Type"] = "application/json";
  this.body = JSON.stringify(object, null, "   ") + "\n";
}
