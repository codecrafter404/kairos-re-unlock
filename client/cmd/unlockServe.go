/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/codecrafter404/kairos-re-unlock/common"
	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// unlockServeCmd represents the unlockServe command
var unlockServeCmd = &cobra.Command{
	Use:   "unlock-serve",
	Short: "Serve a http server for the target device",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("unlock password: ")
		var password string
		_, err := fmt.Scanln(&password)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to read input")
		}

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
		ntp, err := cmd.Flags().GetString("ntp")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get ntp server")
		}

		log.Info().Msg("[⏰] Get current time")
		offset := common.QueryOffset(config.Config{
			NTPServer: ntp,
		})

		port, err := cmd.Flags().GetInt("port")
		fmt.Printf("%d\n", port)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get port number")
		}

		mux := http.NewServeMux()
		srv := http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		}
		mux.HandleFunc("/get_payload", func(w http.ResponseWriter, r *http.Request) {
			log.Info().Str("ip", r.RemoteAddr).Msg("[⚒️] Preparing payload")
			payload, err := common.NewPayload(string(publicKey), string(privateKey), password, offset)
			if err != nil {
				log.Error().Err(err).Msg("Failed to generate payload")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			payload_string, err := json.Marshal(payload)
			if err != nil {
				log.Error().Err(err).Msg("Failed to marshall payload")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Write(payload_string)

			go srv.Shutdown(context.Background())
		})

		log.Info().Int("port", port).Msg("[📋] Listening for connections")
		err = srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Str("listen", fmt.Sprintf(":%d", port)).Msg("Failed to listen on port")
		}

	},
}

func init() {
	rootCmd.AddCommand(unlockServeCmd)

	unlockServeCmd.Flags().StringP("public-key", "d", "", "eg ./droplet_pub.pem")
	unlockServeCmd.Flags().StringP("private-key", "c", "", "eg ./client_priv.pem")
	unlockServeCmd.Flags().IntP("port", "p", 505, "the listen port")
	unlockServeCmd.Flags().StringP("ntp", "", "time.cloudflare.com", "The ntp pool used for timestamp setting")
	unlockServeCmd.MarkFlagsOneRequired("public-key", "private-key")
}
