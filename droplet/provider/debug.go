package provider

import (
	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/mudler/go-pluggable"
)

func getDebugPassword(config config.Config, datachan chan<- pluggable.EventResponse) {
	if config.IsDebugEnabled() && config.DebugConfig.Password != "" {
		datachan <- pluggable.EventResponse{
			Data:  config.DebugConfig.Password,
			Error: "",
		}
	}
}
