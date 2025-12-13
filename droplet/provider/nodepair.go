package provider

import (
	"context"

	"github.com/codecrafter404/go-nodepair"
	"github.com/codecrafter404/kairos-re-unlock/common"
	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/mudler/go-pluggable"
	"github.com/rs/zerolog/log"
)

func getAsyncNodePairResponse(config config.Config, channel chan<- pluggable.EventResponse) {
	res, err := getResponse(config)
	err_s := ""
	if err != nil {
		err_s = err.Error()
	}

	channel <- pluggable.EventResponse{
		Data:  res,
		Error: err_s,
	}
}

// TODO: add retry logic
func getResponse(config config.Config) (string, error) {
	// edgevpn get payload
	ctx, _ := context.WithCancel(context.Background())
	payload := &common.Payload{}

	log.Info().Msg("Waiting for payload")
	err := nodepair.Receive(ctx, payload, nodepair.WithToken(config.EdgeVPNToken))

	if err != nil {
		log.Err(err).Msg("Failed to receive payload")
		return "", err
	}
	log.Info().Any("payload", payload).Msg("Payload received")

	return payload.GetPassword(config)
}
