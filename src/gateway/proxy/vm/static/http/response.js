var AP = AP || {};
AP.HTTP = AP.HTTP || {};

/**
 * Creates a new HTTP response that can be returned from proxy endpoints.
 *
 * @class
 * @constructor
 */
AP.HTTP.Response = function() {
  /**
   * The HTTP status code of the response.
   * @type {integer}
   */
  this.url = null;

  /**
   * The body of the response.
   * @type {string}
   */
  this.body = null;

  /**
   * The response's headers.
   * @type {Object.<string,string|string[]>}
   */
  this.headers = null;
}

/**
 * Sets the object as the response's body, formatted as JSON.
 *
 * This method automatically sets the 'Content-Type' header
 * to 'application/json' to match.
 *
 * @param {object} object The object to serialize and use as the response body
 */
AP.HTTP.Response.prototype.setJSONBody = function(object) {
  this.headers = this.headers || {};
  this.headers["Content-Type"] = "application/json";
  this.body = JSON.stringify(object);
}

/**
 * Sets the object as the response's body, formatted as pretty-printed JSON.
 *
 * This method automatically sets the 'Content-Type' header
 * to 'application/json' to match.
 *
 * @param {object} object The object to serialize and use as the response body
 */
AP.HTTP.Response.prototype.setJSONBodyPretty = function(object) {
  this.headers = this.headers || {};
  this.headers["Content-Type"] = "application/json";
  this.body = JSON.stringify(object, null, "   ") + "\n";
}
