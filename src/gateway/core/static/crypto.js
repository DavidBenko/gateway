var AP = AP || {};

/**
 * Crypto contains cryptographic functions.
 *
 * @namespace
 */
AP.Crypto = AP.Crypto || {};

// Otto can only set vars in the highest scope in the VM (global, basically).
// So we'll just move it to the proper namespace and then clean the global up.
AP.Crypto.hashPassword = _hashPassword;
delete _hashPassword;

AP.Crypto.hash = _hash;
delete _hash;

AP.Crypto.hashHmac = _hashHmac
delete _hashHmac

AP.Crypto.sign = _sign;
delete _sign;
