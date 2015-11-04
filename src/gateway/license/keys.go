package license

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"gateway/config"
	apcrypto "gateway/crypto"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

const defaultLicenseFileLocation = "license"

var DeveloperVersion = false
var developerVersionAccounts,
	developerVersionUsers,
	developerVersionAPIs,
	developerVersionProxyEndpoints string
var (
	DeveloperVersionAccounts       = 1
	DeveloperVersionUsers          = 1
	DeveloperVersionAPIs           = 1
	DeveloperVersionProxyEndpoints = 5
)

func init() {
	if value, err := strconv.Atoi(developerVersionAccounts); err == nil {
		DeveloperVersionAccounts = value
	}
	if value, err := strconv.Atoi(developerVersionUsers); err == nil {
		DeveloperVersionUsers = value
	}
	if value, err := strconv.Atoi(developerVersionAPIs); err == nil {
		DeveloperVersionAPIs = value
	}
	if value, err := strconv.Atoi(developerVersionProxyEndpoints); err == nil {
		DeveloperVersionProxyEndpoints = value
	}
}

// ValidateForever reads the signed license file at path, and validates it
// immediately, and then again each interval in a separate goroutine.
// Failure to validate is fatal.
func ValidateForever(path string, interval time.Duration) {
	// if no provided license file, then use license at default location of './license'
	var data []byte
	var err error
	// No path specified for the license
	if path == "" {
		// Default to well-known location 'license' in the current working dir
		if data, err = ioutil.ReadFile(defaultLicenseFileLocation); err != nil {
			// No license file present?  No worries, let's default to the developer version
			if os.IsNotExist(err) {
				log.Printf("%s Starting gateway in developer mode", config.System)
				DeveloperVersion = true
				return
			}
			log.Fatalf("%s Unable to read license file at '%s': %v", config.System, defaultLicenseFileLocation, err)
		}
	} else {
		data, err = ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("%s Could not read license at '%s': %v", config.System, path, err)
		}
	}

	signed, err := DeserializeSignedLicense(data)
	if err != nil {
		log.Fatalf("%s Could not deserialize license at '%s'", config.System, path)
	}

	publicKeyData, err := Asset("public_key")
	if err != nil {
		log.Fatalf("%s Could not find embedded key", config.System)
	}

	pemData, _ := apcrypto.PEMDataFromData(publicKeyData, "PUBLIC KEY")
	publicKey, err := apcrypto.RSAPublicKey(pemData)
	if err != nil {
		log.Fatalf("%s Could not decode embedded key", config.System)
	}

	checkValidity := func() {
		if !signed.IsValid(publicKey) {
			log.Fatalf("%s License at '%s' is not valid.", config.System, path)
		}
	}

	checkValidity()
	go func() {
		for {
			time.Sleep(interval)
			checkValidity()
		}
	}()
}

const signatureHash = 0

// The License interface
type License interface {
	version() int
	valid() bool
}

// A SignedLicense holds all license data including the cryptographic signature.
type SignedLicense struct {
	Data      string
	Version   int
	Signature []byte
}

// License returns the license embedded in Data, based on Version.
func (s *SignedLicense) License() (License, error) {
	v1 := &V1{}
	switch s.Version {
	case v1.version():
		var license V1
		err := json.Unmarshal([]byte(s.Data), &license)
		return &license, err
	}

	return nil, fmt.Errorf("Could not create license from version %v: %v",
		s.Version, s.Data)
}

// IsValid returns true if the license is good enough to keep running Gateway.
func (s *SignedLicense) IsValid(key *rsa.PublicKey) bool {
	license, err := s.License()
	if err != nil || !license.valid() {
		return false
	}
	err = rsa.VerifyPKCS1v15(key, signatureHash, []byte(s.Data), s.Signature)
	return (err == nil)
}

// Serialize returns the json formatted signed license, suitable for writing
// to a file and Deserialize ing in the future.
func (s *SignedLicense) Serialize() ([]byte, error) {
	return json.MarshalIndent(s, "", "    ")
}

// DeserializeSignedLicense deserializes a JSON formatted signed license
func DeserializeSignedLicense(data []byte) (*SignedLicense, error) {
	var license SignedLicense
	err := json.Unmarshal(data, &license)
	return &license, err
}

// GenerateSignedLicense returns a SignedLicense built from the license and key.
func GenerateSignedLicense(license License, key *rsa.PrivateKey) (*SignedLicense, error) {
	licenseJSON, err := json.Marshal(license)
	if err != nil {
		return nil, err
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, key, signatureHash, licenseJSON)
	if err != nil {
		return nil, err
	}

	return &SignedLicense{
		Data:      string(licenseJSON),
		Version:   license.version(),
		Signature: signature,
	}, nil
}
