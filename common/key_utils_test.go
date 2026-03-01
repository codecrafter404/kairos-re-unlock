package common

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func generateTestKeyPair(t *testing.T) (string, string) {
	t.Helper()
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		t.Fatalf("Failed to marshal private key: %v", err)
	}
	pemPriv := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}))

	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}
	pemPub := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))

	return pemPriv, pemPub
}

func TestParsePublicKey(t *testing.T) {
	_, pemPub := generateTestKeyPair(t)
	pubKey, err := ParsePublicKey(pemPub)
	if err != nil {
		t.Fatalf("ParsePublicKey failed: %v", err)
	}
	if pubKey.N == nil {
		t.Fatal("Public key N is nil")
	}
}

func TestParsePublicKey_InvalidPEM(t *testing.T) {
	_, err := ParsePublicKey("not a pem block")
	if err == nil {
		t.Fatal("Expected error for invalid PEM")
	}
}

func TestParsePublicKey_WrongType(t *testing.T) {
	pemPriv, _ := generateTestKeyPair(t)
	_, err := ParsePublicKey(pemPriv)
	if err == nil {
		t.Fatal("Expected error for private key passed as public")
	}
}

func TestParsePrivateKey(t *testing.T) {
	pemPriv, _ := generateTestKeyPair(t)
	privKey, err := ParsePrivateKey(pemPriv)
	if err != nil {
		t.Fatalf("ParsePrivateKey failed: %v", err)
	}
	if privKey.D == nil {
		t.Fatal("Private key D is nil")
	}
}

func TestParsePrivateKey_InvalidPEM(t *testing.T) {
	_, err := ParsePrivateKey("not a pem block")
	if err == nil {
		t.Fatal("Expected error for invalid PEM")
	}
}

func TestParsePrivateKey_WrongType(t *testing.T) {
	_, pemPub := generateTestKeyPair(t)
	_, err := ParsePrivateKey(pemPub)
	if err == nil {
		t.Fatal("Expected error for public key passed as private")
	}
}
