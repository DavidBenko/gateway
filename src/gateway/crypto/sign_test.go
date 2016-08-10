package crypto_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	gCrypto "gateway/crypto"
	"testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func TestSign(t *testing.T) { gc.TestingT(t) }

type SignSuite struct{}

var _ = gc.Suite(&SignSuite{})

var data = "some secret data"

func (s *SignSuite) TestSignAndVerifyRsa(c *gc.C) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	c.Assert(err, jc.ErrorIsNil)

	for i, t := range []struct {
		should      string
		given       []byte
		key         *rsa.PrivateKey
		algo        string
		padding     string
		expectError string
	}{{
		should:  "work for md5 and pkcs1v15",
		given:   []byte(data),
		key:     key,
		algo:    "md5",
		padding: "pkcs1v15",
	}, {
		should:  "work for md5 and pss",
		given:   []byte(data),
		key:     key,
		algo:    "md5",
		padding: "pss",
	}, {
		should:  "work for sha1 and pkcs1v15",
		given:   []byte(data),
		key:     key,
		algo:    "sha1",
		padding: "pkcs1v15",
	}, {
		should:  "work for sha1 and pss",
		given:   []byte(data),
		key:     key,
		algo:    "sha1",
		padding: "pss",
	}, {
		should:  "work for sha256 and pkcs1v15",
		given:   []byte(data),
		key:     key,
		algo:    "sha256",
		padding: "pkcs1v15",
	}, {
		should:  "work for sha256 and pss",
		given:   []byte(data),
		key:     key,
		algo:    "sha256",
		padding: "pss",
	}, {
		should:  "work for sha384 and pkcs1v15",
		given:   []byte(data),
		key:     key,
		algo:    "sha384",
		padding: "pkcs1v15",
	}, {
		should:  "work for sha384 and pss",
		given:   []byte(data),
		key:     key,
		algo:    "sha384",
		padding: "pss",
	}, {
		should:  "work for sha512 and pkcs1v15",
		given:   []byte(data),
		key:     key,
		algo:    "sha512",
		padding: "pkcs1v15",
	}, {
		should:  "work for sha512 and pss",
		given:   []byte(data),
		key:     key,
		algo:    "sha512",
		padding: "pss",
	}, {
		should:      "return an error if supplied invalid padding scheme",
		given:       []byte(data),
		key:         key,
		algo:        "sha256",
		padding:     "foobar",
		expectError: "invalid padding scheme",
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		result, err := gCrypto.Sign(t.given, t.key, t.algo, t.padding)

		if t.expectError != "" {
			c.Assert(err.Error(), gc.Equals, t.expectError)
			continue
		}

		c.Assert(err, jc.ErrorIsNil)
		c.Assert(result, gc.NotNil)

		sig := result.(*gCrypto.RsaSignature)

		valid, err := gCrypto.Verify(t.given, sig.Signature, t.key.Public(), t.algo, t.padding)

		c.Assert(err, jc.ErrorIsNil)
		c.Assert(valid, jc.IsTrue)
	}
}
func (s *SignSuite) TestSignAndVerifyEcdsa(c *gc.C) {
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)

	c.Assert(err, jc.ErrorIsNil)

	for i, t := range []struct {
		should      string
		given       []byte
		key         *ecdsa.PrivateKey
		algo        string
		expectError string
	}{{
		should: "work for md5",
		given:  []byte(data),
		key:    privateKey,
		algo:   "md5",
	}, {
		should: "work for sha1",
		given:  []byte(data),
		key:    privateKey,
		algo:   "sha1",
	}, {
		should: "work for sha256",
		given:  []byte(data),
		key:    privateKey,
		algo:   "sha256",
	}, {
		should: "work for sha512",
		given:  []byte(data),
		key:    privateKey,
		algo:   "sha512",
	}, {
		should: "work for sha384",
		given:  []byte(data),
		key:    privateKey,
		algo:   "sha384",
	}, {
		should: "work for sha512_256",
		given:  []byte(data),
		key:    privateKey,
		algo:   "sha512_256",
	}, {
		should: "work for sha3_224",
		given:  []byte(data),
		key:    privateKey,
		algo:   "sha3_224",
	}, {
		should: "work for sha3_256",
		given:  []byte(data),
		key:    privateKey,
		algo:   "sha3_256",
	}, {
		should: "work for sha3_384",
		given:  []byte(data),
		key:    privateKey,
		algo:   "sha3_384",
	}, {
		should: "work for sha3_512",
		given:  []byte(data),
		key:    privateKey,
		algo:   "sha3_512",
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		result, err := gCrypto.Sign(t.given, t.key, t.algo, "")

		if t.expectError != "" {
			c.Assert(err.Error(), gc.Equals, t.expectError)
			continue
		}

		c.Assert(err, jc.ErrorIsNil)
		c.Assert(result, gc.NotNil)

		signature := result.(*gCrypto.EcdsaSignature)

		c.Assert(signature.Signature, gc.Not(gc.Equals), "")

		valid, err := gCrypto.Verify(t.given, signature.Signature, t.key.Public(), t.algo, "")

		c.Assert(err, jc.ErrorIsNil)
		c.Assert(valid, jc.IsTrue)
	}
}
