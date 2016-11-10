package crypto_test

import (
	"testing"

	gCrypto "gateway/crypto"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func TestAes(t *testing.T) { gc.TestingT(t) }

type AesSuite struct{}

var _ = gc.Suite(&AesSuite{})

func (s *AesSuite) TestParseAesKey(c *gc.C) {
	for i, t := range []struct {
		should            string
		expectedBlockSize int
		expectedKeySize   int
		givenKey          string
		givenIV           string
		expectError       string
	}{{
		should:            "make a 16 byte block",
		expectedBlockSize: 16,
		expectedKeySize:   16,
		givenKey:          "aaaabbbbccccdddd",
		givenIV:           "aaaabbbbccccdddd",
	}, {
		should:            "make a 32 byte key with a 16 byte block",
		expectedBlockSize: 16,
		expectedKeySize:   32,
		givenKey:          "aaaabbbbccccddddeeeeffffgggghhhh",
		givenIV:           "aaaabbbbccccdddd",
	}, {
		should:      "return an error if key is invalid size; should be 16, 24 or 32 bytes",
		givenKey:    "not valid",
		givenIV:     "this is some ivs",
		expectError: "crypto/aes: invalid key size 9",
	}, {
		should:      "return an error if the key is too large",
		givenKey:    "aaaabbbbccccddddeeeeffffgggghhhhaaaabbbbccccddddeeeeffffgggghhhh",
		givenIV:     "aaaabbbbccccddddeeeeffffgggghhhhaaaabbbbccccddddeeeeffffgggghhhh",
		expectError: "crypto/aes: invalid key size 64",
	}, {
		should:      "return an error if key is empty",
		givenKey:    "",
		givenIV:     "aaaabbbbccccdddd",
		expectError: "missing key",
	}, {
		should:      "return an error if IV is longer than key blocksize",
		givenKey:    "aaaabbbbccccdddd",
		givenIV:     "aaaabbbbccccddddeeeeffffgggghhhh",
		expectError: "iv should be 16 bytes, got 32 bytes",
	}, {
		should:            "not return an error if the IV is empty",
		expectedBlockSize: 16,
		expectedKeySize:   16,
		givenKey:          "aaaabbbbccccdddd",
		givenIV:           "", // supplying no IV will generate one during encryption
		expectError:       "",
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		key, err := gCrypto.ParseAesKey([]byte(t.givenKey), []byte(t.givenIV), gCrypto.CFBMode)

		if t.expectError != "" {
			c.Assert(err.Error(), gc.Equals, t.expectError)
			continue
		}

		c.Assert(err, jc.ErrorIsNil)
		c.Assert(key, gc.NotNil)
		c.Assert(key.Block.BlockSize(), gc.Equals, t.expectedBlockSize)
		c.Assert(len(key.Key), gc.Equals, t.expectedKeySize)
	}
}

func (s *AesSuite) TestEncryptAes(c *gc.C) {
	for i, t := range []struct {
		should      string
		given       string
		givenKey    string
		givenIV     string
		expectError string
		mode        gCrypto.AesMode
	}{{
		should:   "work with 16 byte key and supplied IV - CFB",
		given:    "foobar16",
		givenKey: "aaaabbbbccccdddd",
		givenIV:  "0000000000000001",
		mode:     gCrypto.CFBMode,
	}, {
		should:   "work with 24 byte key and supplied IV - CFB",
		given:    "foobar24",
		givenKey: "aaaabbbbccccddddeeeeffff",
		givenIV:  "0000000000000001",
		mode:     gCrypto.CFBMode,
	}, {
		should:   "work with 32 byte key and supplied IV - CFB",
		given:    "foobar32",
		givenKey: "aaaabbbbccccddddeeeeffffgggghhhh",
		givenIV:  "fafafa1234ddeegg",
		mode:     gCrypto.CFBMode,
	}, {
		should:   "work without a supplied IV, one should be generated - CFB",
		given:    "foobar",
		givenKey: "aaaabbbbccccddddeeeeffffgggghhhh",
		mode:     gCrypto.CFBMode,
	}, {
		should:   "work with 16 byte key and supplied IV - CBC",
		given:    "foobar16",
		givenKey: "aaaabbbbccccdddd",
		givenIV:  "0000000000000001",
		mode:     gCrypto.CBCMode,
	}, {
		should:   "work with 24 byte key and supplied IV - CBC",
		given:    "foobar24",
		givenKey: "aaaabbbbccccddddeeeeffff",
		givenIV:  "0000000000000001",
		mode:     gCrypto.CBCMode,
	}, {
		should:   "work with 32 byte key and supplied IV - CBC",
		given:    "foobar24",
		givenKey: "aaaabbbbccccddddeeeeffffgggghhhh",
		givenIV:  "0000000000000001",
		mode:     gCrypto.CBCMode,
	}, {
		should:   "work without a supplied IV, one should be generated - CBC",
		given:    "foobar",
		givenKey: "aaaabbbbccccddddeeeeffffgggghhhh",
		mode:     gCrypto.CBCMode,
	}} {
		c.Logf("Test %d: should %s", i, t.should)

		key, err := gCrypto.ParseAesKey([]byte(t.givenKey), []byte(t.givenIV), t.mode)

		c.Assert(err, jc.ErrorIsNil)
		c.Assert(key, gc.NotNil)

		result, err := gCrypto.EncryptAes([]byte(t.given), key)
		if t.expectError != "" {
			c.Assert(err.Error(), gc.Equals, t.expectError)
			continue
		}
		c.Assert(err, jc.ErrorIsNil)

		back, err := gCrypto.DecryptAes(result, key)
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(string(back), gc.Equals, t.given)
	}
}
