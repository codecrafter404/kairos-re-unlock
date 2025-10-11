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
		privKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to generate private key")
		}
		pubKey := privKey.Public()

		publicBlock, err := x509.MarshalPKIXPublicKey(pubKey)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to marshal public key")
		}
		pemPublicKey := string(pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicBlock,
		}))

		privateBlock, err := x509.MarshalPKCS8PrivateKey(privKey)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to marshal private key")
		}
		pemPrivateKey := string(pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privateBlock,
		}))
		token := nodepair.GenerateToken()

		fmt.Printf(`kcrypt
   remote_unlock:
      edgevpn_token: %s
      public_key: |
%s
      private_key: |
%s
`, token, padLeft(pemPublicKey), padLeft(pemPrivateKey))
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
