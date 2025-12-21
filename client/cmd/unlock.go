/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/codecrafter404/go-nodepair"
	"github.com/codecrafter404/kairos-re-unlock/common"
	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// unlockCmd represents the unlock command
var unlockCmd = &cobra.Command{
	Use:   "unlock [password]",
	Short: "Unlock target device (nodepair)",
	Long:  `Sends the encrypted and singed payload to the pair in order to let them decrypt their drive`,
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
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

		token, err := cmd.Flags().GetString("edgevpn-token")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get token")
		}

		ntp, err := cmd.Flags().GetString("ntp")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get ntp server")
		}

		log.Info().Msg("[⏰] Get current time")
		offset := common.QueryOffset(config.Config{
			NTPServer: ntp,
		})
		log.Info().Msg("[⚒️] Preparing payload")
		payload, err := common.NewPayload(string(publicKey), string(privateKey), password, offset)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to generate payload")
		}

		log.Info().Msg("[📬] Sending payload")

		err = nodepair.Send(cmd.Context(), payload, nodepair.WithToken(token))
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to send payload")
		}

		log.Info().Msg("[🏁] Sucessfully sent unlock payload")
	},
}

func init() {
	rootCmd.AddCommand(unlockCmd)

	unlockCmd.Flags().StringP("public-key", "d", "", "eg ./droplet_pub.pem")
	unlockCmd.Flags().StringP("private-key", "c", "", "eg ./client_priv.pem")
	unlockCmd.Flags().StringP("edgevpn-token", "e", "", "The EdgeVPN token")
	unlockCmd.Flags().StringP("ntp", "", "de.pool.ntp.org", "The ntp pool used for timestamp setting")

	unlockCmd.MarkFlagsOneRequired("public-key", "private-key", "edgevpn-token")
}
