package crypto_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"testing"

	gCrypto "gateway/crypto"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func TestEncryption(t *testing.T) { gc.TestingT(t) }

type EncryptionSuite struct{}

var _ = gc.Suite(&EncryptionSuite{})

func (s *EncryptionSuite) TestEncryption(c *gc.C) {
	data := "something that should be encrypted"
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)

	c.Assert(err, jc.ErrorIsNil)
	for i, t := range []struct {
		should      string
		given       string
		key         *rsa.PrivateKey
		expectError string
	}{{
		should: "work for rsa key",
		given:  data,
		key:    privateKey,
	}} {
		c.Logf("Test %d: should %s", i, t.should)

		result, err := gCrypto.Encrypt([]byte(data), t.key.Public(), "sha256", "")

		if t.expectError != "" {
			c.Assert(err.Error(), gc.Equals, t.expectError)
			continue
		}

		c.Assert(err, jc.ErrorIsNil)
		c.Assert(result, gc.NotNil)

		decodedResult, _ := base64.StdEncoding.DecodeString(result)

		decryptedResult, err := gCrypto.Decrypt([]byte(decodedResult), t.key, "sha256", "")

		c.Assert(err, jc.ErrorIsNil)
		c.Assert(decryptedResult, gc.Equals, t.given)
	}
}

func (s *EncryptionSuite) TestEcdsaEncrypt(c *gc.C) {
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)

	c.Assert(err, jc.ErrorIsNil)

	for i, t := range []struct {
		should      string
		key         *ecdsa.PrivateKey
		expectError string
	}{{
		should:      "return an error",
		key:         privateKey,
		expectError: "ECDSA should only be used for signing, not encryption",
	}} {
		c.Logf("Test %d: should %s", i, t.should)

		_, err := gCrypto.Encrypt([]byte("secret"), t.key.Public(), "sha256", "")

		if t.expectError != "" {
			c.Assert(err.Error(), gc.Equals, t.expectError)
			continue
		}
	}
}
