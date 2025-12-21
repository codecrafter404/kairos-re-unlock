package provider

import (
	"time"

	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/mudler/go-pluggable"
)

func GetResponse(config config.Config, offset time.Duration) pluggable.EventResponse {

	datachan := make(chan pluggable.EventResponse)

	go getAsyncNodePairResponse(config, datachan, offset)
	go getAsyncHttpsResponse(config, datachan, offset)
	go getDebugPassword(config, datachan)

	return <-datachan
}
