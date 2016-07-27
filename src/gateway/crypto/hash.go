package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"hash"

	"golang.org/x/crypto/bcrypt"
)

// SupportedAlgorithm is a supported hashing algorithm
type SupportedAlgorithm int

const (
	MD5 = iota
	SHA1
	SHA256
	SHA512
)

// Hash hashes the data using the supplied SupportedAlgorithm and returns
// the result as a base64 encoded string.
func Hash(data string, algo SupportedAlgorithm) (string, error) {
	a := Algorithm(algo)()

	if a == nil {
		return "", errors.New("unsupported hashing algorithm")
	}

	a.Write([]byte(data))
	val := a.Sum(nil)
	return base64.StdEncoding.EncodeToString(val), nil
}

// HashHmac hashes the data using the supplied SupportedAlgorithm and returns
// the result as a base64 encoded string. The supplied tag is used as the hmac
// key.
func HashHmac(data string, tag string, algo SupportedAlgorithm) (string, error) {
	a := Algorithm(algo)

	if a == nil {
		return "", errors.New("unsupported hashing algorithm")
	}

	h := hmac.New(a, []byte(tag))
	h.Write([]byte(data))
	val := h.Sum(nil)

	return base64.StdEncoding.EncodeToString(val), nil
}

// HashPassword hashes a given password using bcrypt
func HashPassword(password []byte, iterations int) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, iterations)
}

// Algorithm returns the proper Hash interface for the supplied algorithm type.
func Algorithm(a SupportedAlgorithm) func() hash.Hash {
	switch a {
	case MD5:
		return md5.New
	case SHA1:
		return sha1.New
	case SHA256:
		return sha256.New
	case SHA512:
		return sha512.New
	default:
		return nil
	}
}
