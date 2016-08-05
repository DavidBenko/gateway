package crypto

import (
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Hash hashes the data using the supplied SupportedAlgorithm and returns
// the result as a base64 encoded string.
func Hash(data string, algo string) (string, error) {
	a, err := GetSupportedAlgorithm(algo)

	if err != nil {
		return "", err
	}

	algorithm := a.New()

	algorithm.Write([]byte(data))
	val := algorithm.Sum(nil)
	return base64.StdEncoding.EncodeToString(val), nil
}

// HashHmac hashes the data using the supplied SupportedAlgorithm and returns
// the result as a base64 encoded string. The supplied tag is used as the HMAC
// key.
func HashHmac(data string, tag string, algo string) (string, error) {
	a, err := GetSupportedAlgorithm(algo)

	if err != nil {
		return "", err
	}

	h := hmac.New(a.New, []byte(tag))
	h.Write([]byte(data))
	val := h.Sum(nil)

	return base64.StdEncoding.EncodeToString(val), nil
}

// HashPassword hashes a given password using bcrypt.
func HashPassword(password string, iterations int) (string, error) {
	result, err := bcrypt.GenerateFromPassword([]byte(password), iterations)

	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(result), nil
}

func CompareHashAndPassword(hash, password string) (bool, error) {
	decoded, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(decoded), []byte(password))

	return err == nil, err
}

// GetSupportedAlgorithm returns the SupportedAlgorithm corresponding to the
// string representation supplied in the javascript.
func GetSupportedAlgorithm(algorithm string) (crypto.Hash, error) {
	name := strings.ToLower(algorithm)

	switch name {
	case "md5":
		return crypto.MD5, nil
	case "sha1":
		return crypto.SHA1, nil
	case "sha256":
		return crypto.SHA256, nil
	case "sha512":
		return crypto.SHA512, nil
	case "sha384":
		return crypto.SHA384, nil
	case "sha512_256":
		return crypto.SHA512_256, nil
	case "sha3_224":
		return crypto.SHA3_224, nil
	case "sha3_256":
		return crypto.SHA3_256, nil
	case "sha3_384":
		return crypto.SHA3_384, nil
	case "sha3_512":
		return crypto.SHA3_512, nil
	default:
		return crypto.MD5, errors.New("unsupported hashing algorithm")
	}
}
