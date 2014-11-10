var AP = AP || {};

/**
 * Creates a new HTTP session, or returns an existing one if already created.
 *
 * @class
 * @constructor
 */
AP.Session = function(id) {
  /**
   * The id of the session. This is used as the name of the cookie.
   * @type {string}
   */
  this.id = id;
}

/**
 * Get a value from the session.
 *
 * @param {string} key The key under which the value was stored.
 */
AP.Session.prototype.get = function(key) {
  return __ap_session_get(this.id, key);
}

/**
* Set a value in the session.
*
* @param {string} key The key under which to store the value.
* @param {string|number|boolean} value The value to store.
*/
AP.Session.prototype.set = function(key, value) {
  __ap_session_set(this.id, key, value);
}

/**
* Checks whether a given key has a value stored against it in this session.
*
* @param {string} key The key to check.
* @return {boolean} Whether or not the key has a value
*/
AP.Session.prototype.isSet = function(key) {
  return __ap_session_is_set(this.id, key);
}

/**
 * Deletes the value from the key in the session.
 *
 * @param {string} key The key to remove.
 */
AP.Session.prototype.delete = function(key) {
  __ap_session_delete(this.id, key);
}

/**
 * Set various options on the session.
 *
 * Options may include the following keys:
 *
 * * path: The path for which to apply the cookie
 * * domain: The domain for which to apply the cookie
 * * maxAge: Use 0 or omit to leave cookie's 'Max-Age' unspecified,
 *           use a value < 0 to set the cookie 'Max-Age: 0',
 *           and use a value > 0 to set the max age to that number of seconds.
 * * secure: Indicate whether the cookie's secure flag should be set.
 * * httpOnly: Indicate whether the cookie should be marked http only.
 *
 * @param {object} options The session cookie's options
 */
AP.Session.prototype.setOptions = function(options) {
  __ap_session_set_options(this.id, JSON.stringify(options));
}
