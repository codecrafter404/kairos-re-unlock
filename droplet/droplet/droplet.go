package droplet

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/codecrafter404/kairos-re-unlock/droplet/provider"
	"github.com/rs/zerolog/log"
)

func Start(config config.Config, offset time.Duration) error {
	log.Debug().Any("config", config).Msg("Starting discovery")

	// capture input
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	res := provider.GetResponse(config, offset)

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old
	log.Info().Str("output", string(out)).Msg("Got Nodepair logs")

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	err := enc.Encode(res)

	return err
}
