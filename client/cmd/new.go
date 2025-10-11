/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"slices"
	"strings"

	"github.com/codecrafter404/go-nodepair"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Initalize a new client config",
	Run: func(cmd *cobra.Command, args []string) {
		// dropleet
		priv1, pub1, err := generateKeyPair()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to generate first keypair")
		}
		// client
		priv2, pub2, err := generateKeyPair()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to generate second keypair")
		}

		token := nodepair.GenerateToken()

		fmt.Printf(`======[Droplet Configuration]======

kcrypt:
   remote_unlock:
      edgevpn_token: %s
      # Public Key of the client
      public_key: |
%s
      # Private Key of Droplet
      private_key: |
%s

======[client_priv.pem]======
%s
======[droplet_pub.pem]======
%s
`, token, padLeft(pub2), padLeft(priv1), priv2, pub1)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func padLeft(data string) string {
	lines := strings.Split(data, "\n")
	lines = slices.DeleteFunc(lines, func(line string) bool { return line == "" })
	res := make([]string, len(lines))
	for i, line := range lines {
		res[i] = fmt.Sprintf("         %s", line)
	}
	return strings.Join(res, "\n")
}

func generateKeyPair() (string, string, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("Failed to generate private key: %w", err)
	}
	pubKey := privKey.Public()

	publicBlock, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", "", fmt.Errorf("Failed to marshal public key: %w", err)
	}
	pemPublicKey := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicBlock,
	}))

	privateBlock, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return "", "", fmt.Errorf("Failed to marshal private key: %w", err)
	}
	pemPrivateKey := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateBlock,
	}))
	return pemPrivateKey, pemPublicKey, nil
}
