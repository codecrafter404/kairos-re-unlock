/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/codecrafter404/kairos-re-unlock/common"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get the logs from the client",
	Run: func(cmd *cobra.Command, args []string) {
		host, err := cmd.Flags().GetIP("ip")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get ip")
		}
		req_url := url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:505", host.String()),
			Path:   "/logs",
		}
		resp, err := http.Get(req_url.String())
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to send response")
		}
		defer resp.Body.Close()

		resp_bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to read response body")
		}

		if resp.StatusCode != http.StatusOK {
			log.Fatal().Err(errors.New(string(resp_bytes))).Msg("Got error back")
		}
		path, err := cmd.Flags().GetString("private-key")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get path")
		}
		privateKey, err := os.ReadFile(path)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to read privateKey")
		}
		//
		var logsPayload common.LogsPayload
		err = json.Unmarshal(resp_bytes, &logsPayload)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to unmarshall logs")
		}

		privKey, err := common.ParsePrivateKey(string(privateKey))
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to parse private key")
		}
		decodedPassword, err := base64.StdEncoding.DecodeString(logsPayload.Key)
		if err != nil {
			log.Fatal().Err(err).Str("base64", string(resp_bytes)).Msg("Failed to decode password")
		}
		aesKey, err := rsa.DecryptOAEP(crypto.SHA512.New(), rand.Reader, &privKey, decodedPassword, nil)
		if err != nil {
			log.Fatal().Err(err).Msg("OAEP decryption failed")
		}

		block, err := aes.NewCipher(aesKey)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create new cipher")
		}
		aesGCM, err := cipher.NewGCM(block)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create new gcm")
		}
		nonce, err := base64.RawStdEncoding.DecodeString(logsPayload.Nonce)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to decode nonce")
		}
		cyphertext, err := base64.RawStdEncoding.DecodeString(logsPayload.Data)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to decode ciphertext")
		}
		logs, err := aesGCM.Open(nonce, nonce, cyphertext, nil)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to unseal logs")
		}
		fmt.Println(string(logs))

	},
}

func init() {
	rootCmd.AddCommand(logsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// logsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// logsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	logsCmd.Flags().IPP("ip", "i", net.ParseIP("127.0.0.1"), "eg. 127.0.0.1")
	logsCmd.Flags().StringP("private-key", "c", "", "eg ./client_priv.pem")
	logsCmd.MarkFlagsOneRequired("ip", "private-key")
}
