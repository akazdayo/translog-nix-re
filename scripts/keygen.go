package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"

	"golang.org/x/mod/sumdb/note"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run ./scripts/keygen.go <signer-name>")
		os.Exit(2)
	}

	privateKey, publicKey, err := note.GenerateKey(rand.Reader, os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if _, err := note.NewVerifier(publicKey); err != nil {
		log.Fatalf("invalid signer name %q: %v", os.Args[1], err)
	}

	// stdout can be captured directly as LOG_PRIVATE_KEY.
	if _, err := fmt.Fprintln(os.Stdout, privateKey); err != nil {
		log.Fatal(err)
	}
	if _, err := fmt.Fprintln(os.Stderr, publicKey); err != nil {
		log.Fatal(err)
	}
}
