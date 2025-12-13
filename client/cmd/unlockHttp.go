/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
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

// unlockHttpCmd represents the unlockHttp command
var unlockHttpCmd = &cobra.Command{
	Use:   "unlock-http [password]",
	Short: "Unlock target device over http",
	Long:  `Sends the encrypted and singed payload (using http) to the pair in order to let them decrypt their drive`,
	Run: func(cmd *cobra.Command, args []string) {
		password := args[0]

		log.Info().Msg("[🏗️] Starting up")

		path, err := cmd.Flags().GetString("public-key")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get path")
		}
		publicKey, err := os.ReadFile(path)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to read publicKey")
		}

		path, err = cmd.Flags().GetString("private-key")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get path")
		}
		privateKey, err := os.ReadFile(path)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to read privateKey")
		}

		log.Info().Msg("[⚒️] Preparing payload")
		payload, err := common.NewPayload(string(publicKey), string(privateKey), password)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to generate payload")
		}

		log.Info().Msg("[👨‍⚕️] Check if system is healthy")
		host, err := cmd.Flags().GetIP("ip")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get ip")
		}
		req_url := url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:505", host.String()),
			Path:   "/health",
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
		if resp.StatusCode != http.StatusOK || string(resp_bytes) != "healthy" {
			log.Fatal().Err(err).Msg("Droplet is unhealthy")
		}

		log.Info().Msg("[📫] Sending payload")

		req_url = url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:505", host.String()),
			Path:   "/unlock",
		}
		payload_string, err := json.Marshal(payload)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to marshal json payload")
		}

		resp, err = http.Post(req_url.String(), "application/json", bytes.NewBuffer(payload_string))

		defer resp.Body.Close()

		resp_bytes, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to read response body")
		}

		if resp.StatusCode != http.StatusAccepted {
			log.Fatal().Int("statuscode", resp.StatusCode).Msg("Got invalid status code")
		} else if string(resp_bytes) != "received" {
			log.Fatal().Str("resp", string(resp_bytes)).Msg("Got invalid response body")
		}

		log.Info().Msg("[🏁] Sucessfully sent unlock payload")

	},
}

func init() {
	rootCmd.AddCommand(unlockHttpCmd)

	unlockHttpCmd.Flags().StringP("public-key", "d", "", "eg ./droplet_pub.pem")
	unlockHttpCmd.Flags().StringP("private-key", "c", "", "eg ./client_priv.pem")
	unlockHttpCmd.Flags().IPP("ip", "i", net.ParseIP("127.0.0.1"), "eg. 127.0.0.1")

	unlockHttpCmd.MarkFlagsOneRequired("public-key", "private-key", "ip")
}
