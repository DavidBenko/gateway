var AP = AP || {};
AP.HTTP = AP.HTTP || {};

AP.HTTP.Request = function() {
  this.__ap_type = "HTTP";

  this.method = "GET";
  this.url = null;
  this.body = null;
  this.headers = {};
}
