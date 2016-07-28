package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"gateway/logreport"
)

// Sign signs the data using the privKey and algorithm.
func Sign(data []byte, privKey interface{}, algorithmName string) ([]byte, error) {
	hash, err := GetSupportedAlgorithm(algorithmName)

	if err != nil {
		return nil, err
	}

	a := hash.New()
	a.Write(data)
	hashed := a.Sum(nil)

	switch privKey.(type) {
	case *rsa.PrivateKey:
		return privKey.(*rsa.PrivateKey).Sign(rand.Reader, hashed, hash)
	case *ecdsa.PrivateKey:
		return privKey.(*ecdsa.PrivateKey).Sign(rand.Reader, hashed, hash)
	default:
		logreport.Println("invalid or unsupported private key type")
		return nil, errors.New("invalid or unsupported private key type")
	}
}
