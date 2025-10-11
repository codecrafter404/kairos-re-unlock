package droplet

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

type ReceivedPayload struct {
	EncyptedPassword string
	Timestamp        int64
	Signature        string
}

func (r ReceivedPayload) hash() [64]byte {
	return sha512.Sum512([]byte(fmt.Sprintf("%s%d", r.EncyptedPassword, r.Timestamp)))
}

// / returns `nil` if the signature is valid
func (r ReceivedPayload) IsValidSignature(publicKey string) error {
	publicBlock, _ := pem.Decode([]byte(publicKey))
	if publicBlock == nil || publicBlock.Type != "PUBLIC KEY" {
		return fmt.Errorf("key provided is not a public key")
	}

	pubKeyAny, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		return fmt.Errorf("Failed to parse public key: %w", err)
	}

	if _, ok := pubKeyAny.(*rsa.PublicKey); !ok {
		return fmt.Errorf("Invalid key type")
	}
	pubKey, _ := pubKeyAny.(*rsa.PublicKey)

	signatureDecoded, err := base64.StdEncoding.DecodeString(r.Signature)
	if err != nil {
		return fmt.Errorf("Failed to decode signature: %w", err)
	}

	hash := r.hash()
	err = rsa.VerifyPSS(pubKey, crypto.SHA512, hash[:], signatureDecoded, nil)
	if err != nil {
		return fmt.Errorf("Invalid signature: %w", err)
	}

	return nil
}

func (r ReceivedPayload) DecryptPassword(privateKey string) (string, error) {
	privateBlock, _ := pem.Decode([]byte(privateKey))
	if privateBlock == nil || privateBlock.Type != "PRIVATE KEY" {
		return "", fmt.Errorf("Key is not a private key")
	}

	privKeyAny, err := x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
	if err != nil {
		return "", fmt.Errorf("Failed to parse private key block: %w", err)
	}
	if _, ok := privKeyAny.(*rsa.PrivateKey); !ok {
		return "", fmt.Errorf("Invalid key type")
	}
	privkey, _ := privKeyAny.(*rsa.PrivateKey)

	decodedPassword, err := base64.StdEncoding.DecodeString(r.EncyptedPassword)
	if err != nil {
		return "", fmt.Errorf("Failed to decode password")
	}

	data, err := rsa.DecryptOAEP(crypto.SHA512.New(), rand.Reader, privkey, decodedPassword, nil)
	if err != nil {
		return "", fmt.Errorf("OAEP decryption failed: %w", err)
	}

	return string(data), nil
}
