package main

import (
	"log"
	"os"
)

func main() {
	usage := "Please run with version of key you wish to generate, e.g.\n"
	usage += "    keygen v1 <options>"

	if len(os.Args) < 2 {
		log.Fatal(usage)
	}

	switch os.Args[1] {
	case "v1":
		generateV1()
	default:
		log.Fatal(usage)
	}
}
