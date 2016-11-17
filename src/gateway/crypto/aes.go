package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

type SymmetricKey struct {
	Key   []byte
	IV    []byte
	Block cipher.Block
	Mode  AesMode
}

type AesMode int

const (
	CFBMode = iota
	CBCMode
)

func pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error, probably incorrect key or AES mode")
	}

	return src[:(length - unpadding)], nil
}

func EncryptAes(data []byte, key *SymmetricKey) ([]byte, error) {
	switch key.Mode {
	case CFBMode:
		return encryptCFB(data, key), nil
	case CBCMode:
		return encryptCBC(data, key), nil
	default:
		return nil, errors.New("invalid AES encryption mode")
	}
}

func encryptCFB(data []byte, key *SymmetricKey) []byte {
	data = pad(data)
	ciphertext := append([]byte(key.IV), make([]byte, len(data))...)

	cfb := cipher.NewCFBEncrypter(key.Block, ciphertext[:key.Block.BlockSize()])
	cfb.XORKeyStream(ciphertext[key.Block.BlockSize():], data)

	return ciphertext
}

func encryptCBC(data []byte, key *SymmetricKey) []byte {
	data = pad(data)
	ciphertext := append([]byte(key.IV), make([]byte, len(data))...)

	cbc := cipher.NewCBCEncrypter(key.Block, key.IV)
	cbc.CryptBlocks(ciphertext[key.Block.BlockSize():], data)

	return ciphertext
}

func DecryptAes(data []byte, key *SymmetricKey) ([]byte, error) {
	switch key.Mode {
	case CFBMode:
		return decryptCFB(data, key)
	case CBCMode:
		return decryptCBC(data, key)
	default:
		return nil, errors.New("invalid AES decryption mode")
	}
}

func decryptCFB(data []byte, key *SymmetricKey) ([]byte, error) {
	d := data[key.Block.BlockSize():]

	cfb := cipher.NewCFBDecrypter(key.Block, data[:key.Block.BlockSize()])
	cfb.XORKeyStream(d, d)

	d, err := unpad(d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func decryptCBC(data []byte, key *SymmetricKey) ([]byte, error) {
	d := data[key.Block.BlockSize():]

	cbc := cipher.NewCBCDecrypter(key.Block, data[:key.Block.BlockSize()])
	cbc.CryptBlocks(d, d)

	d, err := unpad(d)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func ParseAesKey(k []byte, i []byte, mode AesMode) (*SymmetricKey, error) {
	if len(k) == 0 {
		return nil, errors.New("missing key")
	}

	block, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}

	if len(i) > 0 && len(i) != aes.BlockSize {
		return nil, errors.New(fmt.Sprintf("iv should be %d bytes, got %d bytes", aes.BlockSize, len(i)))
	}

	symkey := &SymmetricKey{Key: k, IV: i, Block: block, Mode: mode}
	if len(symkey.IV) == 0 {
		iv := make([]byte, aes.BlockSize)
		_, err := rand.Read(iv)
		if err != nil {
			return nil, err
		}
		symkey.IV = iv
	}

	return symkey, nil
}
