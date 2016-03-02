var AP = AP || {};

/**
 * Creates a new Call object that can hold a request and a response.
 *
 * @class
 * @constructor
 */
AP.Call = function() {
  /**
   * The request object to use
   * @type {Object}
   */
  this.request = null;

  /**
   * The response from the request.
   * @type {Object}
   */
  this.response = null;
}
