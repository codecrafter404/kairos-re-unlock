package droplet

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/codecrafter404/kairos-re-unlock/droplet/notify"
	"github.com/codecrafter404/kairos-re-unlock/droplet/provider"
	"github.com/mudler/go-pluggable"
	"github.com/rs/zerolog/log"
)

func Start(config config.Config, offset time.Duration) error {
	log.Debug().Any("config", config).Msg("Starting discovery")

	// capture input
	var res *pluggable.EventResponse
	for {

		log.Info().Msg("Waiting for unlock password")
		notify.SendNotification("waiting for unlock password", config)
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		maybe_response := provider.GetResponse(config, offset)
		if maybe_response != nil {
			res = maybe_response
		}

		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = old
		log.Info().Str("output", string(out)).Msg("Got Nodepair logs")

		if res != nil {
			break
		}
		log.Info().Msg("As we got no response, we sleep for 30s")
		notify.SendNotification("Got no valid password; timeout for 30s", config)
		time.Sleep(time.Second * 30)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	err := enc.Encode(*res)
	notify.SendNotification("successfull unlock", config)

	return err
}
