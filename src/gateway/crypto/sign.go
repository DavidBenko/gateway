package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	_ "crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"

	_ "golang.org/x/crypto/sha3"
)

type EcdsaSignature struct {
	R         *big.Int
	S         *big.Int
	Signature string
}

type RsaSignature struct {
	Signature string
}

// Sign signs the data using the privKey and algorithm.
func Sign(data []byte, privKey interface{}, algorithmName string, padding string) (interface{}, error) {
	hash, err := GetSupportedAlgorithm(algorithmName)

	if err != nil {
		return nil, err
	}

	a := hash.New()
	a.Write(data)
	hashed := a.Sum(nil)

	switch privKey.(type) {
	case *rsa.PrivateKey:
		switch strings.ToLower(padding) {
		case "pkcs1v15":
			r, err := rsa.SignPKCS1v15(rand.Reader, privKey.(*rsa.PrivateKey), hash, hashed[:])
			if err != nil {
				return nil, err
			}

			sig := &RsaSignature{base64.StdEncoding.EncodeToString(r)}
			return sig, nil
		case "pss":
			r, err := rsa.SignPSS(rand.Reader, privKey.(*rsa.PrivateKey), hash, hashed[:], nil)
			if err != nil {
				return nil, err
			}

			sig := &RsaSignature{base64.StdEncoding.EncodeToString(r)}
			return sig, nil
		default:
			return nil, errors.New("invalid padding scheme")
		}
	case *ecdsa.PrivateKey:
		// ECDSA does not support/require a padding scheme.
		r, s, err := ecdsa.Sign(rand.Reader, privKey.(*ecdsa.PrivateKey), hashed[:])
		if err != nil {
			return nil, err
		}

		signature := r.Bytes()
		signature = append(signature, s.Bytes()...)

		sig := &EcdsaSignature{R: r, S: s, Signature: base64.StdEncoding.EncodeToString(signature)}
		return sig, nil
	default:
		return nil, errors.New(fmt.Sprintf("invalid or unsupported private key type: %T", privKey))
	}
}

func Verify(data []byte, signature interface{}, publicKey interface{}, algorithmName string, padding string) (bool, error) {
	hash, err := GetSupportedAlgorithm(algorithmName)

	if err != nil {
		return false, err
	}

	a := hash.New()
	a.Write(data)
	hashed := a.Sum(nil)

	switch publicKey.(type) {
	case *rsa.PublicKey:
		switch strings.ToLower(padding) {
		case "pkcs1v15":
			sig := signature.(*RsaSignature)
			decodedSignature, err := base64.StdEncoding.DecodeString(sig.Signature)

			if err != nil {
				return false, err
			}

			err = rsa.VerifyPKCS1v15(publicKey.(*rsa.PublicKey), hash, hashed[:], decodedSignature)
			return err == nil, err
		case "pss":
			sig := signature.(*RsaSignature)
			decodedSignature, err := base64.StdEncoding.DecodeString(sig.Signature)

			if err != nil {
				return false, err
			}

			err = rsa.VerifyPSS(publicKey.(*rsa.PublicKey), hash, hashed[:], decodedSignature, nil)
			return err == nil, err
		default:
			return false, errors.New("invalid or unsupported padding scheme")
		}
	case *ecdsa.PublicKey:
		// ECDSA does not support/require a padding scheme.
		sig := signature.(*EcdsaSignature)
		valid := ecdsa.Verify(publicKey.(*ecdsa.PublicKey), hashed, sig.R, sig.S)
		return valid, nil
	default:
		return false, errors.New(fmt.Sprintf("invalid or unsupported public key type: %T", publicKey))
	}
}
