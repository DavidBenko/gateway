var AP = AP || {};

AP.Session = function(id) {
  this.id = id;
}

AP.Session.prototype.get = function(key) {
  return __ap_session_get(this.id, key);
}

AP.Session.prototype.set = function(key, value) {
  __ap_session_set(this.id, key, value);
}

AP.Session.prototype.isSet = function(key) {
  return __ap_session_is_set(this.id, key);
}

AP.Session.prototype.delete = function(key) {
  __ap_session_delete(this.id, key);
}

AP.Session.prototype.setOptions = function(options) {
  __ap_session_set_options(this.id, JSON.stringify(options));
}
