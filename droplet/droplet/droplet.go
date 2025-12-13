package droplet

import (
	"os"

	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/codecrafter404/kairos-re-unlock/droplet/provider"
	"github.com/kairos-io/kcrypt/pkg/bus"
	"github.com/mudler/go-pluggable"
	"github.com/rs/zerolog/log"
)

func Start(config config.Config) error {
	log.Debug().Any("config", config).Msg("Starting discovery")
	factory := pluggable.NewPluginFactory()

	// Input: bus.EventInstallPayload
	// Expected output: map[string]string{}
	factory.Add(bus.EventDiscoveryPassword, func(e *pluggable.Event) pluggable.EventResponse {
		return provider.GetResponse(config)
	})
	return factory.Run(pluggable.EventType(os.Args[1]), os.Stdin, os.Stdout)
}
