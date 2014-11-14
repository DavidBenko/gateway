package crypto

import (
	"crypto/rsa"
	"crypto/x509"
)

// RSAPrivateKeyAtPath returns the RSA Private Key at the PEM-encoded path
func RSAPrivateKeyAtPath(path string) (*rsa.PrivateKey, error) {
	data, err := PEMDataFromPath(path, "RSA PRIVATE KEY")
	if err != nil {
		return nil, err
	}

	return x509.ParsePKCS1PrivateKey(data)
}

// RSAPublicKeyAtPath returns the RSA Public Key at the PEM-encoded path
func RSAPublicKeyAtPath(path string) (*rsa.PublicKey, error) {
	data, err := PEMDataFromPath(path, "PUBLIC KEY")
	if err != nil {
		return nil, err
	}
	return RSAPublicKey(data)
}

// RSAPublicKey returns the RSA public key from the PEM-encoded data
func RSAPublicKey(data []byte) (*rsa.PublicKey, error) {
	key, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return nil, err
	}

	return key.(*rsa.PublicKey), nil
}
