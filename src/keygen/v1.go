package main

import (
	"flag"
	"fmt"
	apcrypto "gateway/crypto"
	"gateway/license"
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

	l := &license.V1{
		Name:       *name,
		Company:    *company,
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
