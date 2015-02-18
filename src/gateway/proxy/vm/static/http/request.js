var AP = AP || {};

/**
 * HTTP holds helper classes for HTTP related tasks, such as making
 * requests with Gateway, and parsing their responses.
 *
 * @namespace
 */
AP.HTTP = AP.HTTP || {};

/**
 * Creates a new HTTP request that can be handed to {@link AP.makeRequest}.
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
 * Sets the object as the request's body, formatted as JSON.
 *
 * This method automatically sets the 'Content-Type' header
 * to 'application/json' to match.
 *
 * @param {object} object The object to serialize and use as the request body
 */
AP.HTTP.Request.prototype.setJSONBody = function(object) {
  this.headers["Content-Type"] = "application/json";
  this.body = JSON.stringify(object);
}
