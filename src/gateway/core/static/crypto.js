var AP = AP || {};

/**
 * Crypto contains cryptographic functions.
 *
 * @namespace
 */
AP.Crypto = AP.Crypto || {};

AP.Crypto.PaddingSchemes = {
    pkcs:   "pkcs1v15",
    pss:        "pss",
}

AP.Crypto.HashingAlgorithms = {
    md5:        "md5",
    sha1:       "sha1",
    sha256:     "sha256",
    sha512:     "sha512",
    sha384:     "sha384",
    sha512_256: "sha512_256",
    sha3_224:   "sha3_224",
    sha3_256:   "sha3_256",
    sha3_384:   "sha3_384",
    sha3_512:   "sha3_512",
}

AP.Crypto.hashPassword = _hashPassword;
delete _hashPassword;

AP.Crypto.compareHashAndPassword = _compareHashAndPassword;
delete _compareHashAndPassword;

AP.Crypto.hash = _hash;
delete _hash;

AP.Crypto.hashHmac = _hashHmac
delete _hashHmac;

