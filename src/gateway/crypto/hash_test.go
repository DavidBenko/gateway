package crypto_test

import (
	gCrypto "gateway/crypto"
	"testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func TestHash(t *testing.T) { gc.TestingT(t) }

type HashSuite struct{}

var _ = gc.Suite(&HashSuite{})

func (s *CryptoSuite) TestHash(c *gc.C) {
	for i, t := range []struct {
		should      string
		given       string
		expect      string
		algo        string
		expectError string
	}{{
		should: "work for MD5",
		given:  "secret",
		algo:   "md5",
		expect: "Xr4ilOzQ4PCOq3aQ0qbuaQ==",
	}, {
		should: "work for sha1",
		given:  "secret",
		algo:   "sha1",
		expect: "5en6G6MezRroT3XKqkdPOmY/BfQ=",
	}, {
		should: "work for sha256",
		given:  "secret",
		algo:   "sha256",
		expect: "K7gNU3sdo+OL0wNhqoVWhr3g6s1xYv72ol/pe/Unols=",
	}, {
		should: "work for sha512",
		given:  "secret",
		algo:   "sha512",
		expect: "vSsar3708Jvp9Szi2NWZZ02Bqp1qRCFpbcTZPdBhnWgs5WtNZKnvCXdhztmeD2cmW192CF5bDufKRpayrW/isg==",
	}, {
		should: "work for sha384",
		given:  "secret",
		algo:   "sha384",
		expect: "WKd1ukESvjAFrkQHznV9iP2nHUBJe7gCbsrFTU4//HIyzo3jq1rLMK45dg/ufFPt",
	}, {
		should: "work for sha512_256",
		given:  "secret",
		algo:   "sha512_256",
		expect: "RAfLu+yTYJcF2DYIs3kNo4N7OmGdxzNG0QthM9zrXrc=",
	}, {
		should: "work for sha3_224",
		given:  "secret",
		algo:   "sha3_256",
		expect: "9aUgeocpsfcJy3EDEXUesvyKytWh+4rJkbc25ptlKaM=",
	}, {
		should: "work for sha3_256",
		given:  "secret",
		algo:   "sha3_256",
		expect: "9aUgeocpsfcJy3EDEXUesvyKytWh+4rJkbc25ptlKaM=",
	}, {
		should: "work for sha3_384",
		given:  "secret",
		algo:   "sha3_384",
		expect: "UiLduG1gYdLA7yu8YHJx/281XUKD/VQmd2a4juGGypOrDkIfMUJ1XVb3buh4icuM",
	}, {
		should: "work for sha3_512",
		given:  "secret",
		algo:   "sha3_512",
		expect: "t3ijmjZjcZ38XkjJ14QxseRcKvnfU4eCvxmcGJ2r6sdoCtpX3OyO7pHE4787+pr2/96QzR0knRxhIde3WaABsQ==",
	}, {
		should:      "return error for unsupported hashing algorithm",
		algo:        "foobar_256",
		expectError: "unsupported hashing algorithm",
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		result, err := gCrypto.Hash(t.given, t.algo)

		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			continue
		}

		c.Assert(err, jc.ErrorIsNil)
		c.Check(result, gc.Equals, t.expect)
	}
}

func (s *CryptoSuite) TestHashHmac(c *gc.C) {
	for i, t := range []struct {
		should      string
		given       string
		expect      string
		algo        string
		expectError string
	}{{
		should: "work for md5",
		given:  "secret",
		algo:   "md5",
		expect: "WOB+qBnvwkpALX5otwkvUg==",
	}, {
		should: "work for sha1",
		given:  "secret",
		algo:   "sha1",
		expect: "7gmdd5WgGjI4mdkYfmyBpS5tltE=",
	}, {
		should: "work for sha256",
		given:  "secret",
		algo:   "sha256",
		expect: "Jk81TozzSCouFN8MJ9xmXcI4UW6UbZzjIVWzQ+7kEl8=",
	}, {
		should: "work for sha512",
		given:  "secret",
		algo:   "sha512",
		expect: "/AxKciWoMHRX8VvzB6o2H3TtLVPw/cxBw0v4LPDYsbT6MuMwrW/Xdtwy3Oizf1wuQP1DvjHs2EnssAkCWSJsWw==",
	}, {
		should: "work for sha384",
		given:  "secret",
		algo:   "sha384",
		expect: "gv1yIky1VJaGGDb7D7lT6pYluJbhPjQ1qLdsA6YBYnRatliG882JnBNXs+VksM9G",
	}, {
		should: "work for sha512_256",
		given:  "secret",
		algo:   "sha512_256",
		expect: "ZfwfwLYHVr/9oa2M93rp/++fv+TKJRq6JBPtqI5fQ9U=",
	}, {
		should: "work for sha3_224",
		given:  "secret",
		algo:   "sha3_256",
		expect: "DDrLfAM8/zhy1yk5vJF4woDtf8DW8mJwizfag0MxEok=",
	}, {
		should: "work for sha3_256",
		given:  "secret",
		algo:   "sha3_256",
		expect: "DDrLfAM8/zhy1yk5vJF4woDtf8DW8mJwizfag0MxEok=",
	}, {
		should: "work for sha3_384",
		given:  "secret",
		algo:   "sha3_384",
		expect: "hCav2zHdE4xvYNT+kSREjL21oxKj1bObULKUppJD9um2R6bB8XH0UXOHbdLsYCBf",
	}, {
		should: "work for sha3_512",
		given:  "secret",
		algo:   "sha3_512",
		expect: "1ZH0goRFeEG2kkkPxt4XYhmxYsn81wh2FSFydjZAqbdUWiAK4Py1Yar/i17pgJ3dNHWMV8pUaR+uidjGKvTxAQ==",
	}, {
		should:      "return error for unsupported hashing algorithm",
		algo:        "foobar_256",
		expectError: "unsupported hashing algorithm",
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		tag := "tests"
		result, err := gCrypto.HashHmac(t.given, tag, t.algo)

		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			continue
		}

		c.Assert(err, jc.ErrorIsNil)
		c.Check(result, gc.Equals, t.expect)

	}
}

func (s *CryptoSuite) TestHashAndComparePassword(c *gc.C) {
	for i, t := range []struct {
		should     string
		given      string
		iterations int
	}{{
		should:     "return a hashed password",
		given:      "s3cr3t",
		iterations: 4,
	}, {
		should:     "return and check a hashed password with more iterations",
		given:      "super_s3cr3t",
		iterations: 10,
	}} {
		c.Logf("Test %d: should %s", i, t.should)

		result, err := gCrypto.HashPassword([]byte(t.given), t.iterations)

		c.Assert(err, jc.ErrorIsNil)
		c.Check(result, gc.Not(gc.Equals), "")

		valid, err := gCrypto.CompareHashAndPassword(result, t.given)

		c.Assert(err, jc.ErrorIsNil)
		c.Assert(valid, jc.IsTrue)
	}
}
