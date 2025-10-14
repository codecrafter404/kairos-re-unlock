package droplet

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/codecrafter404/go-nodepair"
	"github.com/codecrafter404/kairos-re-unlock/common"
	"github.com/kairos-io/kcrypt/pkg/bus"
	"github.com/mudler/go-pluggable"
	"github.com/rs/zerolog/log"
)

const cache_path = "/var/log/kairos/kairos-re-unlock-cache"

func Start(config Config) error {
	log.Debug().Any("config", config).Msg("Starting discovery")
	factory := pluggable.NewPluginFactory()

	// Input: bus.EventInstallPayload
	// Expected output: map[string]string{}
	factory.Add(bus.EventDiscoveryPassword, func(e *pluggable.Event) pluggable.EventResponse {

		// var data_s string
		// var err_s string
		//
		// cache, err := os.ReadFile(cache_path)
		//
		// if err != nil {
		// 	log.Info().Err(err).Msg("Cache miss")
		// 	data_s, err = getResponse(config)
		// 	if err != nil {
		// 		err_s = err.Error()
		// 	} else {
		// 		// cache
		// 		err := os.WriteFile(cache_path, []byte(data_s), 0600)
		// 		if err != nil {
		// 			log.Err(err).Msg("Could not write cache file")
		// 		}
		// 	}
		// } else {
		// 	log.Info().Msg("Cache hit")
		// 	data_s = string(cache)
		// }
		//
		// // respond with password
		// return pluggable.EventResponse{
		// 	Data:  data_s,
		// 	Error: err_s,
		// }
		return pluggable.EventResponse{
			Data:  "password",
			Error: "",
		}
	})
	return factory.Run(pluggable.EventType(os.Args[1]), os.Stdin, os.Stdout)
}

// TODO: add retry logic
func getResponse(config Config) (string, error) {
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

	// check signature
	err = payload.IsValidSignature(config.PublicKey)

	if err != nil {
		log.Err(err).Msg("Signature verification failed")
		return "", err
	}

	// check timestamp
	now := time.Now().Unix()
	var gracePeriod int64 = 60 * 15

	if payload.Timestamp < (now-gracePeriod) || payload.Timestamp > now {
		err = fmt.Errorf("Invalid timestamp")
		log.Err(err).Int64("grace", gracePeriod).Int64("now", now).Msg("Invalid timestamp")
		return "", err
	}

	// decrypt payload
	password, err := payload.DecryptPassword(config.PrivateKey)
	if err != nil {
		log.Err(err).Msg("Decryption failed")
		return "", nil
	}

	return password, nil
}
