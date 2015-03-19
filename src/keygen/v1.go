package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	apcrypto "gateway/crypto"
	"gateway/license"
	"io"
	"log"
	"os"
	"time"
)

func generateV1() {
	v1 := flag.NewFlagSet("v1", flag.ContinueOnError)
	name := v1.String("name", "", "Name of registered user")
	company := v1.String("company", "", "Company of registered user")
	expires := v1.Int64("expires-in", -1, "Number of days license is valid, or -1 for perpetual")
	keyPath := v1.String("private-key", "", "Path to the RSA private key")
	v1.Parse(os.Args[2:])

	if *name == "" || *keyPath == "" {
		log.Fatal("Must invoke with at least name and path to private key.")
	}

	key, err := apcrypto.RSAPrivateKeyAtPath(*keyPath)
	if err != nil {
		log.Fatal(err)
	}

	var expiration *time.Time
	if *expires > -1 {
		day := 24 * time.Hour
		days := time.Duration(*expires) * day
		future := time.Now().Add(days)
		expiration = &future
	}

	uuid, err := newUUID()
	if err != nil {
		log.Fatal(err)
	}

	l := &license.V1{
		Name:       *name,
		Company:    *company,
		Id:         uuid,
		Expiration: expiration,
	}

	signed, err := license.GenerateSignedLicense(l, key)
	if err != nil {
		log.Fatal(err)
	}

	serialized, err := signed.Serialize()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v\n", string(serialized))
}

// newUUID generates a random UUID according to RFC 4122
func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}

	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40

	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8],
		uuid[8:10], uuid[10:]), nil
}
