package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"gateway/config"
	"gateway/logreport"
)

func Encrypt(data []byte, publicKey interface{}, algorithmName string, tag string) (string, error) {
	hash, err := GetSupportedAlgorithm(algorithmName)

	if err != nil {
		return "", err
	}

	if hash == crypto.MD5 || hash == crypto.SHA1 {
		logreport.Printf("%s Minimum of SHA256 is recommended.\n", config.Admin)
	}

	switch publicKey.(type) {
	case *rsa.PublicKey:
		result, err := rsa.EncryptOAEP(hash.New(), rand.Reader, publicKey.(*rsa.PublicKey), data, []byte(tag))
		if err != nil {
			logreport.Print(err)
			return "", err
		}

		return base64.StdEncoding.EncodeToString(result), nil
	case *ecdsa.PublicKey:
		return "", errors.New("ECDSA should only be used for signing, not encryption")
	default:
		return "", errors.New(fmt.Sprintf("invalid or unsupported public key type %T", publicKey))
	}
}

func Decrypt(data []byte, privateKey interface{}, algorithmName string, tag string) (string, error) {
	hash, err := GetSupportedAlgorithm(algorithmName)

	if err != nil {
		return "", err
	}

	switch privateKey.(type) {
	case *rsa.PrivateKey:
		if err != nil {
			return "", err
		}
		result, err := rsa.DecryptOAEP(hash.New(), rand.Reader, privateKey.(*rsa.PrivateKey), data, []byte(tag))
		if err != nil {
			return "", err
		}

		return string(result), nil
	case *ecdsa.PrivateKey:
		return "", errors.New("ECDSA should only be used for signing, not encryption")
	default:
		return "", errors.New(fmt.Sprintf("invalid or unsupported private key type %T", privateKey))
	}
}
