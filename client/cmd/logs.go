/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"io"
	"net"
	"net/http"
	"net/url"
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
		fmt.Println(string(resp_bytes))

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
	logsCmd.MarkFlagsOneRequired("ip")
}
