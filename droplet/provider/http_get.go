package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/codecrafter404/kairos-re-unlock/common"
	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/mudler/go-pluggable"
	"github.com/rs/zerolog/log"
)

func getHttpPullResponse(config config.Config, channel chan<- pluggable.EventResponse, offset time.Duration, ctx context.Context) {
	res := make(chan pluggable.EventResponse)

	go func() {
		if len(config.HttpPull) == 0 {
			return
		}
		for {
			for _, address := range config.HttpPull {
				if ctx.Err() != nil {
					return
				}
				if !strings.Contains(address, ":") {
					address += ":505"
				}
				log.Info().Str("addr", address).Msg("Trying to pull payload")

				url := url.URL{
					Scheme: "http",
					Host:   address,
					Path:   "get_payload",
				}
				client := *http.DefaultClient
				client.Timeout = 5 * time.Second
				http_res, err := client.Get(url.String())
				if err != nil {
					log.Error().Err(err).Str("addr", url.String()).Msg("Failed to send off request")
					continue
				}
				defer http_res.Body.Close()

				var passwd common.Payload

				decoder := json.NewDecoder(http_res.Body)

				err = decoder.Decode(&passwd)
				if err != nil {
					log.Err(err).Msg("Failed to decode json body")
					continue
				}

				password, err := passwd.GetPassword(config, offset)

				if err != nil {
					log.Err(err).Msg("Failed to get password")
					continue
				}

				res <- pluggable.EventResponse{
					Data:  password,
					Error: "",
				}
			}

		}
	}()

	select {
	case x := <-res:
		channel <- x
	case <-ctx.Done():
	}
}
