package droplet

import (
	"context"
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

func Start(config config.Config, offset time.Duration, stdin []byte) error {
	log.Debug().Any("config", config).Msg("Starting discovery")

	// capture input
	var res *pluggable.EventResponse
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	for {

		log.Info().Msg("Waiting for unlock password")
		notify.SendNotification("waiting for unlock password", config)

		ctx, cancel := context.WithCancel(context.Background())

		maybe_response := provider.GetResponse(config, offset, stdin, ctx)
		if maybe_response != nil {
			res = maybe_response
		}

		cancel()

		if res != nil {
			break
		}
		log.Info().Msg("As we got no response, we sleep for 30s")
		notify.SendNotification("Got no valid password; timeout for 30s", config)
		time.Sleep(time.Second * 30)
	}

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old
	log.Info().Str("output", string(out)).Msg("Got Nodepair logs")

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	err := enc.Encode(*res)
	notify.SendNotification("successfull unlock", config)

	return err
}

// StartUnlock waits for a password via the configured providers and returns
// the raw password string. If device is non-empty, the password is validated
// against that LUKS device before being returned.
func StartUnlock(config config.Config, offset time.Duration, device string) (string, error) {
	log.Debug().Any("config", config).Msg("Starting unlock")

	// Build stdin payload for password validation
	var stdin []byte
	if device != "" {
		stdin = []byte(`{"data":"{\"name\":\"` + device + `\"}"}`)
	} else {
		stdin = []byte("{}")
	}

	// capture stdout from nodepair
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var res *pluggable.EventResponse
	for {
		log.Info().Msg("Waiting for unlock password")
		notify.SendNotification("waiting for unlock password", config)

		ctx, cancel := context.WithCancel(context.Background())

		maybe_response := provider.GetResponse(config, offset, stdin, ctx)
		if maybe_response != nil {
			res = maybe_response
		}

		cancel()

		if res != nil {
			break
		}
		log.Info().Msg("As we got no response, we sleep for 30s")
		notify.SendNotification("Got no valid password; timeout for 30s", config)
		time.Sleep(time.Second * 30)
	}

	w.Close()
	_, _ = io.ReadAll(r) // drain pipe; errors are non-fatal
	os.Stdout = old

	notify.SendNotification("successful unlock", config)

	return res.Data, nil
}
