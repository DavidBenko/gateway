var AP = AP || {};

/**
 * HTTP holds helper classes for HTTP related tasks, such as making
 * requests with Gateway, and parsing their responses.
 *
 * @namespace
 */
AP.HTTP = AP.HTTP || {};

/**
 * Creates a new HTTP request.
 *
 * @class
 * @constructor
 * @param [request] - An incoming request to copy the method, query, headers, and body.
 */
AP.HTTP.Request = function() {
  /**
   * The HTTP method to use.
   * @type {string}
   */
  this.method = "GET";

  /**
   * The body of the request.
   * @type {string}
   */
  this.body = null;

  /**
   * The request's query parameters.
   * @type {Object.<string,string>}
   */
  this.query = {};

  /**
   * The request's headers.
   * @type {Object.<string,string|string[]>}
   */
  this.headers = {};

  if (arguments.length == 1) {
    var request = arguments[0];
    this.method = _.clone(request.method);
    this.query = _.clone(request.query);
    this.headers = _.clone(request.headers);
    this.body = _.clone(request.body);
  }
}

/**
 * Sets the object as the request's body, formatted as an HTTP encoded form.
 *
 * This method automatically sets the 'Content-Type' header
 * to 'application/json' to match.
 *
 * @param {object} object The object to serialize and use as the request body
 */
AP.HTTP.Request.prototype.setFormBody = function(object) {
  this.method = "POST";
  this.headers["Content-Type"] = "application/x-www-form-urlencoded";
  this.body = AP.HTTP.Request.encodeForm(object);
}

AP.HTTP.Request.encodeForm = function(object) {
  var encodeValue = function(encoded, key, value) {
    if (_.isArray(value)) {
      return encodeArray(encoded, key, value);
    }
    if (_.isObject(value)) {
      return encodeObject(encoded, key, value);
    }
    return key + "=" + encodeURI(value);
  };
  var encodeArray = function(encoded, key, arr) {
    _.map(arr, function(value) {
      if (encoded != "") {
        encoded += "&";
      }
      encoded += encodeValue(encoded, key + "[]", value);
    });
    return encoded;
  };
  var encodeObject = function(encoded, key, obj) {
    _.map(obj, function(value, objKey) {
      if (encoded != "") {
        encoded += "&";
      }
      encoded += encodeValue(encoded, key + "[" + objKey + "]", value);
    });
    return encoded;
  };
  var encoded = "";
  _.map(object, function(v, k) {
    if (encoded != "") {
      encoded += "&";
    }
    encoded += encodeValue("", k, v);
  });
  return encoded;
}
