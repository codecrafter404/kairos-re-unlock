package notify

import (
	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/rs/zerolog/log"
)

func SendNotification(msg string, config config.Config) {
	if config.DiscordWebhook == "" {
		log.Warn().Msg("No discord webhook provided")
		return
	}

	client, err := webhook.NewWithURL(config.DiscordWebhook)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize discord webhook")
	}
	_, err = client.CreateMessage(discord.WebhookMessageCreate{
		Content: msg,
		// Username: "KairosReUnlock",
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send discord message")
	}
}
