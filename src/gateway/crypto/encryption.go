package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
)

func Encrypt(key crypto.PublicKey, data []byte) ([]byte, error) {
	switch key.(type) {
	case *rsa.PublicKey:
		return rsa.EncryptOAEP(sha256.New(), rand.Reader, key.(*rsa.PublicKey), data, []byte{})
	case *ecdsa.PublicKey:
		return nil, errors.New("ECDSA should only be used for signing, not encryption")
	default:
		return nil, errors.New("invalid or unsupported public key type")
	}
}

func Decrypt(key crypto.PrivateKey, data []byte) ([]byte, error) {
	return nil, nil
}
