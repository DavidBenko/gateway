package license

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"testing"
	"time"
)

var (
	privateKey, _ = rsa.GenerateKey(rand.Reader, 1024)
	publicKey     = &privateKey.PublicKey

	future = time.Now().Add(time.Hour)
	past   = time.Now().Add(-time.Hour)

	validLicense   = &V1{Expiration: &future}
	invalidLicense = &V1{Expiration: &past}
)

func TestGenerateSignedLicense(t *testing.T) {
	signed, err := GenerateSignedLicense(validLicense, privateKey)
	if err != nil {
		t.Errorf("Error generating license: %v", err)
		t.Fail()
	}
	embedded, err := signed.License()
	if err != nil {
		t.Errorf("Error getting embedded license: %v", err)
		t.Fail()
	}
	if *embedded.(*V1).Expiration != future {
		t.Error("Expected license in signed version to be the same")
	}
}

func TestSerializeSignedLicense(t *testing.T) {
	signed, _ := GenerateSignedLicense(validLicense, privateKey)
	expected, _ := json.MarshalIndent(signed, "", "    ")
	actual, err := signed.Serialize()
	if err != nil || string(actual) != string(expected) {
		t.Error("Expected serialization format to be pretty JSON")
	}
}

func TestDeserializeSignedLicense(t *testing.T) {
	signed, _ := GenerateSignedLicense(validLicense, privateKey)
	data, _ := signed.Serialize()
	deserialized, err := DeserializeSignedLicense(data)
	if err != nil {
		t.Errorf("Expected to deserialize signed license %v\n", err)
	}
	embedded, _ := deserialized.License()
	if *embedded.(*V1).Expiration != future {
		t.Error("Expected license in signed version to be the same")
	}
}

func TestSignedValid(t *testing.T) {
	signed, _ := GenerateSignedLicense(validLicense, privateKey)
	if !signed.IsValid(publicKey) {
		t.Error("Expected signed license with valid underlying license to be valid")
	}
}

func TestSignedValidUnderlyingInvalid(t *testing.T) {
	signed, _ := GenerateSignedLicense(invalidLicense, privateKey)
	if signed.IsValid(publicKey) {
		t.Error("Expected signed license with invalid underlying license to be invalid")
	}
}

func TestSignedValidBadJSON(t *testing.T) {
	signed, _ := GenerateSignedLicense(validLicense, privateKey)
	signed.Data = "Junk"
	if signed.IsValid(publicKey) {
		t.Error("Expected signed license with junk data to be invalid")
	}
}

func TestSignedValidSignatureMismatch(t *testing.T) {
	wrongKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	signed, _ := GenerateSignedLicense(validLicense, wrongKey)
	if signed.IsValid(publicKey) {
		t.Error("Expected license signed with wrong key to be invalid")
	}
}
