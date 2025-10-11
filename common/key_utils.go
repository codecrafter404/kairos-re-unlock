package common

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func ParsePublicKey(pemPublicKey string) (rsa.PublicKey, error) {
	publicBlock, _ := pem.Decode([]byte(pemPublicKey))
	if publicBlock == nil || publicBlock.Type != "PUBLIC KEY" {
		return rsa.PublicKey{}, fmt.Errorf("key provided is not a public key")
	}

	pubKeyAny, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		return rsa.PublicKey{}, fmt.Errorf("Failed to parse public key: %w", err)
	}

	if _, ok := pubKeyAny.(*rsa.PublicKey); !ok {
		return rsa.PublicKey{}, fmt.Errorf("Invalid key type")
	}
	pubKey, _ := pubKeyAny.(*rsa.PublicKey)
	return *pubKey, nil
}
func ParsePrivateKey(pemPrivateKey string) (rsa.PrivateKey, error) {
	privateBlock, _ := pem.Decode([]byte(pemPrivateKey))
	if privateBlock == nil || privateBlock.Type != "PRIVATE KEY" {
		return rsa.PrivateKey{}, fmt.Errorf("Key is not a private key")
	}

	privKeyAny, err := x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
	if err != nil {
		return rsa.PrivateKey{}, fmt.Errorf("Failed to parse private key block: %w", err)
	}
	if _, ok := privKeyAny.(*rsa.PrivateKey); !ok {
		return rsa.PrivateKey{}, fmt.Errorf("Invalid key type")
	}
	privkey, _ := privKeyAny.(*rsa.PrivateKey)
	return *privkey, nil
}
