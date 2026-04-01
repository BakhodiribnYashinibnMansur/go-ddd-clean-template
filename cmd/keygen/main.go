package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("failed to generate private key: %v\n", err)
		os.Exit(1)
	}

	// Encode private key to PEM (PKCS#1)
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	// Encode public key to PEM (PKIX)
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		fmt.Printf("failed to marshal public key: %v\n", err)
		os.Exit(1)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	// Combine into single PEM file
	combined := append(privPEM, pubPEM...)

	err = os.WriteFile("keypair.pem", combined, 0600)
	if err != nil {
		fmt.Printf("failed to write keypair.pem: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully generated keypair.pem (private + public key)")
	fmt.Println(string(combined))
}
