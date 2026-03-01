package common

import (
	"testing"
	"time"

	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
)

func TestNewPayload_RoundTrip(t *testing.T) {
	// Generate two key pairs: one for droplet, one for client
	dropletPriv, dropletPub := generateTestKeyPair(t)
	clientPriv, clientPub := generateTestKeyPair(t)

	password := "test-luks-passphrase"

	// Client creates a payload encrypted to droplet's public key, signed with client's private key
	payload, err := NewPayload(dropletPub, clientPriv, password, 0)
	if err != nil {
		t.Fatalf("NewPayload failed: %v", err)
	}

	// Verify signature with client's public key
	err = payload.IsValidSignature(clientPub)
	if err != nil {
		t.Fatalf("Signature validation failed: %v", err)
	}

	// Decrypt password with droplet's private key
	decrypted, err := payload.DecryptPassword(dropletPriv)
	if err != nil {
		t.Fatalf("DecryptPassword failed: %v", err)
	}

	if decrypted != password {
		t.Fatalf("Expected password %q, got %q", password, decrypted)
	}
}

func TestNewPayload_InvalidSignatureKey(t *testing.T) {
	dropletPriv, dropletPub := generateTestKeyPair(t)
	clientPriv, _ := generateTestKeyPair(t)
	_, wrongPub := generateTestKeyPair(t) // different key pair

	password := "test-password"

	payload, err := NewPayload(dropletPub, clientPriv, password, 0)
	if err != nil {
		t.Fatalf("NewPayload failed: %v", err)
	}

	// Verify with wrong public key should fail
	err = payload.IsValidSignature(wrongPub)
	if err == nil {
		t.Fatal("Expected signature validation to fail with wrong key")
	}

	// Decrypt with wrong private key should fail
	_, err = payload.DecryptPassword(dropletPriv)
	if err != nil {
		t.Fatalf("DecryptPassword with correct key should work: %v", err)
	}
}

func TestNewPayload_WrongDecryptionKey(t *testing.T) {
	_, dropletPub := generateTestKeyPair(t)
	clientPriv, _ := generateTestKeyPair(t)
	wrongPriv, _ := generateTestKeyPair(t)

	password := "test-password"

	payload, err := NewPayload(dropletPub, clientPriv, password, 0)
	if err != nil {
		t.Fatalf("NewPayload failed: %v", err)
	}

	// Decrypt with wrong private key should fail
	_, err = payload.DecryptPassword(wrongPriv)
	if err == nil {
		t.Fatal("Expected decryption to fail with wrong key")
	}
}

func TestPayload_GetPassword(t *testing.T) {
	dropletPriv, dropletPub := generateTestKeyPair(t)
	clientPriv, clientPub := generateTestKeyPair(t)

	password := "my-luks-password"

	payload, err := NewPayload(dropletPub, clientPriv, password, 0)
	if err != nil {
		t.Fatalf("NewPayload failed: %v", err)
	}

	// Simulate droplet-side config
	conf := config.Config{
		PublicKey:  clientPub,
		PrivateKey: dropletPriv,
	}

	decrypted, err := payload.GetPassword(conf, 0)
	if err != nil {
		t.Fatalf("GetPassword failed: %v", err)
	}
	if decrypted != password {
		t.Fatalf("Expected %q, got %q", password, decrypted)
	}
}

func TestPayload_Valid_ExpiredTimestamp(t *testing.T) {
	dropletPriv, dropletPub := generateTestKeyPair(t)
	clientPriv, clientPub := generateTestKeyPair(t)

	password := "test-password"

	// Create payload with large negative offset to simulate old timestamp
	payload, err := NewPayload(dropletPub, clientPriv, password, -20*time.Minute)
	if err != nil {
		t.Fatalf("NewPayload failed: %v", err)
	}

	conf := config.Config{
		PublicKey:  clientPub,
		PrivateKey: dropletPriv,
	}

	err = payload.Valid(conf, 0)
	if err == nil {
		t.Fatal("Expected timestamp validation to fail for expired payload")
	}
}

func TestPayload_Valid_FutureTimestamp(t *testing.T) {
	dropletPriv, dropletPub := generateTestKeyPair(t)
	clientPriv, clientPub := generateTestKeyPair(t)

	password := "test-password"

	// Create payload with future timestamp (> 0 offset just past grace period)
	payload, err := NewPayload(dropletPub, clientPriv, password, 20*time.Minute)
	if err != nil {
		t.Fatalf("NewPayload failed: %v", err)
	}

	conf := config.Config{
		PublicKey:  clientPub,
		PrivateKey: dropletPriv,
	}

	err = payload.Valid(conf, 0)
	if err == nil {
		t.Fatal("Expected timestamp validation to fail for future payload")
	}
}

func TestPayload_Valid_WithinGracePeriod(t *testing.T) {
	dropletPriv, dropletPub := generateTestKeyPair(t)
	clientPriv, clientPub := generateTestKeyPair(t)

	password := "test-password"

	// Create payload with small offset (within 15 min grace period)
	payload, err := NewPayload(dropletPub, clientPriv, password, -5*time.Minute)
	if err != nil {
		t.Fatalf("NewPayload failed: %v", err)
	}

	conf := config.Config{
		PublicKey:  clientPub,
		PrivateKey: dropletPriv,
	}

	err = payload.Valid(conf, 0)
	if err != nil {
		t.Fatalf("Expected valid payload within grace period, got: %v", err)
	}
}
