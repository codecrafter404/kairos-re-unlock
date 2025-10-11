package common

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"time"
)

type Payload struct {
	EncyptedPassword string
	Timestamp        int64
	Signature        string
}

func (r Payload) hash() [64]byte {
	return sha512.Sum512(fmt.Appendf(nil, "%s%d", r.EncyptedPassword, r.Timestamp))
}

// returns `nil` if the signature is valid
func (r Payload) IsValidSignature(publicKey string) error {
	pubKey, error := ParsePublicKey(publicKey)
	if error != nil {
		return error
	}
	signatureDecoded, err := base64.StdEncoding.DecodeString(r.Signature)
	if err != nil {
		return fmt.Errorf("Failed to decode signature: %w", err)
	}

	hash := r.hash()
	err = rsa.VerifyPSS(&pubKey, crypto.SHA512, hash[:], signatureDecoded, nil)
	if err != nil {
		return fmt.Errorf("Invalid signature: %w", err)
	}

	return nil
}

func (r Payload) DecryptPassword(privateKey string) (string, error) {
	privKey, err := ParsePrivateKey(privateKey)
	if err != nil {
		return "", err
	}
	decodedPassword, err := base64.StdEncoding.DecodeString(r.EncyptedPassword)
	if err != nil {
		return "", fmt.Errorf("Failed to decode password")
	}

	data, err := rsa.DecryptOAEP(crypto.SHA512.New(), rand.Reader, &privKey, decodedPassword, nil)
	if err != nil {
		return "", fmt.Errorf("OAEP decryption failed: %w", err)
	}

	return string(data), nil
}

func (r Payload) signPayload(privateKey string) error {
	privKey, err := ParsePrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("Failed to parse private key: %w", err)
	}
	hash := r.hash()
	signature, err := rsa.SignPSS(rand.Reader, &privKey, crypto.SHA512, hash[:], nil)

	if err != nil {
		return fmt.Errorf("Failed to sign payload: %w", err)
	}
	r.Signature = base64.StdEncoding.EncodeToString(signature)
	return nil
}

func withPassword(pemPublicKey string, password string) (Payload, error) {
	pubKey, err := ParsePublicKey(pemPublicKey)
	if err != nil {
		return Payload{}, fmt.Errorf("Failed to parse public key: %w", err)
	}

	encryptedPassword, err := rsa.EncryptOAEP(crypto.SHA512.New(), rand.Reader, &pubKey, []byte(password), nil)
	if err != nil {
		return Payload{}, fmt.Errorf("Failed to encrypt: %w", err)
	}

	return Payload{
		EncyptedPassword: base64.StdEncoding.EncodeToString(encryptedPassword),
	}, nil
}
func NewPayload(dropletPublicKey string, clientPrivateKey string, password string) (Payload, error) {
	payload, err := withPassword(dropletPublicKey, password)
	if err != nil {
		return Payload{}, err
	}

	payload.Timestamp = time.Now().Unix()

	err = payload.signPayload(clientPrivateKey)
	if err != nil {
		return Payload{}, err
	}

	return payload, nil
}
